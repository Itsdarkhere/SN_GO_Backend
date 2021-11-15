package routes

import (	
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sort"
	"net/http"
	"time"	

	"github.com/deso-protocol/core/lib"
)
type NFTCollectionResponsePlus struct {
	NFTEntryResponse       *NFTEntryResponse     `json:",omitempty"`
	ProfileEntryResponse   *ProfileEntryResponse `json:",omitempty"`
	PostEntryResponse      *PostEntryResponse    `json:",omitempty"`
	HighestBidAmountNanos  uint64                `safeForLogging:"true"`
	LowestBidAmountNanos   uint64                `safeForLogging:"true"`
	NumCopiesForSale       uint64                `safeForLogging:"true"`
	AvailableSerialNumbers []uint64              `safeForLogging:"true"`
}
type GetNFTShowcaseResponsePlus struct {
	NFTCollections []*NFTCollectionResponsePlus
}

func (fes *APIServer) GetNFTShowcasePlus(ww http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(io.LimitReader(req.Body, MaxRequestBodySizeBytes))
	requestData := GetNFTShowcaseRequest{}
	if err := decoder.Decode(&requestData); err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("GetNFTShowcase: Error parsing request body: %v", err))
		return
	}

	var readerPublicKeyBytes []byte
	var err error
	if requestData.ReaderPublicKeyBase58Check != "" {
		readerPublicKeyBytes, _, err = lib.Base58CheckDecode(requestData.ReaderPublicKeyBase58Check)
		if err != nil {
			_AddBadRequestError(ww, fmt.Sprintf("GetNFTShowcase: Problem decoding reader public key: %v", err))
			return
		}
	}

	dropEntry, err := fes.GetLatestNFTDropEntry()
	if err != nil {
		_AddInternalServerError(ww, fmt.Sprintf("GetNFTShowcase: Problem getting latest drop: %v", err))
		return
	}

	currentTime := uint64(time.Now().UnixNano())
	if dropEntry.DropTstampNanos > currentTime {
		// In this case, we have found a pending drop. We must go back one drop in order to
		// get the current active drop.
		if dropEntry.DropNumber == 1 {
			// If the pending drop is drop #1, we need to return a blank dropEntry.
			dropEntry = &NFTDropEntry{}
		}

		if dropEntry.DropNumber > 1 {
			dropNumToFetch := dropEntry.DropNumber - 1
			dropEntry, err = fes.GetNFTDropEntry(dropNumToFetch)
			if err != nil {
				_AddInternalServerError(ww, fmt.Sprintf(
					"GetNFTShowcase: Problem getting drop #%d: %v", dropNumToFetch, err))
				return
			}
		}
	}

	// Now that we have the drop entry, fetch the NFTs.
	utxoView, err := fes.backendServer.GetMempool().GetAugmentedUniversalView()
	if err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("GetNFTShowcase: Error getting utxoView: %v", err))
		return
	}

	var readerPKID *lib.PKID
	if requestData.ReaderPublicKeyBase58Check != "" {
		readerPKID = utxoView.GetPKIDForPublicKey(readerPublicKeyBytes).PKID
	}
	var nftCollectionResponses []*NFTCollectionResponsePlus
	for _, nftHash := range dropEntry.NFTHashes {
		postEntry := utxoView.GetPostEntryForPostHash(nftHash)
		if postEntry == nil {
			_AddInternalServerError(ww, fmt.Sprint("GetNFTShowcase: Found nil post entry for NFT hash."))
			return
		}

		// Should fix the marketplace, stopped working once burn was implemented
		if postEntry.NumNFTCopiesBurned != postEntry.NumNFTCopies {
			nftKey := lib.MakeNFTKey(nftHash, 1)
			nftEntry := utxoView.GetNFTEntryForNFTKey(&nftKey)

			postEntryResponse, err := fes._postEntryToResponse(
				postEntry, false, fes.Params, utxoView, readerPublicKeyBytes, 2)
			if err != nil {
				_AddInternalServerError(ww, fmt.Sprint("GetNFTShowcase: Found invalid post entry for NFT hash."))
				return
			}
			postEntryResponse.PostEntryReaderState = utxoView.GetPostEntryReaderState(readerPublicKeyBytes, postEntry)
			nftCollectionResponsePlus := fes._nftEntryToNFTCollectionResponsePlus(nftEntry, postEntry.PosterPublicKey, postEntryResponse, utxoView, readerPKID)
			nftCollectionResponses = append(nftCollectionResponses, nftCollectionResponsePlus)
		}
	}

	// Return all the data associated with the transaction in the response
	res := GetNFTShowcaseResponsePlus{
		NFTCollections: nftCollectionResponses,
	}

	if err = json.NewEncoder(ww).Encode(res); err != nil {
		_AddInternalServerError(ww, fmt.Sprintf("GetNFTShowcase: Problem serializing object to JSON: %v", err))
		return
	}
}
// Addition to regular showcase request is profilepk, to narrow it to only that profiles nfts
type GetNFTShowcaseRequestProfile struct {
	ReaderPublicKeyBase58Check string `safeForLogging:"true"`
	ProfilePublicKey string `safeForLogging:"true"`
}

func (fes *APIServer) GetNFTShowcaseProfile(ww http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(io.LimitReader(req.Body, MaxRequestBodySizeBytes))
	// Addition
	requestData := GetNFTShowcaseRequestProfile{}
	if err := decoder.Decode(&requestData); err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("GetNFTShowcase: Error parsing request body: %v", err))
		return
	}

	var readerPublicKeyBytes []byte
	var err error
	if requestData.ReaderPublicKeyBase58Check != "" {
		readerPublicKeyBytes, _, err = lib.Base58CheckDecode(requestData.ReaderPublicKeyBase58Check)
		if err != nil {
			_AddBadRequestError(ww, fmt.Sprintf("GetNFTShowcase: Problem decoding reader public key: %v", err))
			return
		}
	}


	dropEntry, err := fes.GetLatestNFTDropEntry()
	if err != nil {
		_AddInternalServerError(ww, fmt.Sprintf("GetNFTShowcase: Problem getting latest drop: %v", err))
		return
	}

	currentTime := uint64(time.Now().UnixNano())
	if dropEntry.DropTstampNanos > currentTime {
		// In this case, we have found a pending drop. We must go back one drop in order to
		// get the current active drop.
		if dropEntry.DropNumber == 1 {
			// If the pending drop is drop #1, we need to return a blank dropEntry.
			dropEntry = &NFTDropEntry{}
		}

		if dropEntry.DropNumber > 1 {
			dropNumToFetch := dropEntry.DropNumber - 1
			dropEntry, err = fes.GetNFTDropEntry(dropNumToFetch)
			if err != nil {
				_AddInternalServerError(ww, fmt.Sprintf(
					"GetNFTShowcase: Problem getting drop #%d: %v", dropNumToFetch, err))
				return
			}
		}
	}

	// Now that we have the drop entry, fetch the NFTs.
	utxoView, err := fes.backendServer.GetMempool().GetAugmentedUniversalView()
	if err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("GetNFTShowcase: Error getting utxoView: %v", err))
		return
	}

	var readerPKID *lib.PKID
	if requestData.ReaderPublicKeyBase58Check != "" {
		readerPKID = utxoView.GetPKIDForPublicKey(readerPublicKeyBytes).PKID
	}

	var nftCollectionResponses []*NFTCollectionResponsePlus
	for _, nftHash := range dropEntry.NFTHashes {
		postEntry := utxoView.GetPostEntryForPostHash(nftHash)
		if postEntry == nil {
			_AddInternalServerError(ww, fmt.Sprint("GetNFTShowcase: Found nil post entry for NFT hash."))
			return
		}

		// Should fix the marketplace, stopped working once burn was implemented
		if postEntry.NumNFTCopiesBurned == postEntry.NumNFTCopies {
			continue;
		}

		nftKey := lib.MakeNFTKey(nftHash, 1)
		nftEntry := utxoView.GetNFTEntryForNFTKey(&nftKey)

		postEntryResponse, err := fes._postEntryToResponse(
			postEntry, false, fes.Params, utxoView, readerPublicKeyBytes, 2)
		if err != nil {
			_AddInternalServerError(ww, fmt.Sprint("GetNFTShowcase: Found invalid post entry for NFT hash."))
			return
		}
		// Addition, this should result in returning only NFTs based on a specific PK
		if postEntryResponse.PosterPublicKeyBase58Check != requestData.ProfilePublicKey {
			continue;
		}
		postEntryResponse.PostEntryReaderState = utxoView.GetPostEntryReaderState(readerPublicKeyBytes, postEntry)
		nftCollectionResponsePlus := fes._nftEntryToNFTCollectionResponsePlus(nftEntry, postEntry.PosterPublicKey, postEntryResponse, utxoView, readerPKID)
		nftCollectionResponses = append(nftCollectionResponses, nftCollectionResponsePlus)
	}

	// Return all the data associated with the transaction in the response
	res := GetNFTShowcaseResponsePlus{
		NFTCollections: nftCollectionResponses,
	}

	if err = json.NewEncoder(ww).Encode(res); err != nil {
		_AddInternalServerError(ww, fmt.Sprintf("GetNFTShowcase: Problem serializing object to JSON: %v", err))
		return
	}
}
func (fes *APIServer) _nftEntryToNFTCollectionResponsePlus(
	nftEntry *lib.NFTEntry,
	posterPublicKey []byte,
	postEntryResponse *PostEntryResponse,
	utxoView *lib.UtxoView,
	readerPKID *lib.PKID,
) *NFTCollectionResponsePlus {

	profileEntry := utxoView.GetProfileEntryForPublicKey(posterPublicKey)
	var profileEntryResponse *ProfileEntryResponse
	if profileEntry != nil {
		profileEntryResponse = fes._profileEntryToResponse(profileEntry, utxoView)
	}

	postEntryResponse.ProfileEntryResponse = profileEntryResponse

	var numCopiesForSale uint64
	serialNumbersForSale := []uint64{}
	for ii := uint64(1); ii <= postEntryResponse.NumNFTCopies; ii++ {
		nftKey := lib.MakeNFTKey(nftEntry.NFTPostHash, ii)
		nftEntryii := utxoView.GetNFTEntryForNFTKey(&nftKey)
		if nftEntryii != nil && nftEntryii.IsForSale {
			if nftEntryii.OwnerPKID != readerPKID {
				serialNumbersForSale = append(serialNumbersForSale, ii)
			}
			numCopiesForSale++
		}
	}
	nftEntryRes := fes._nftEntryToResponse(nftEntry, nil, utxoView, true, readerPKID)

	highestBidAmountNanos, lowestBidAmountNanos := utxoView.GetHighAndLowBidsForNFTCollection(
		nftEntry.NFTPostHash)

	return &NFTCollectionResponsePlus{
		NFTEntryResponse: nftEntryRes,
		ProfileEntryResponse:   profileEntryResponse,
		PostEntryResponse:      postEntryResponse,
		HighestBidAmountNanos:  highestBidAmountNanos,
		LowestBidAmountNanos:   lowestBidAmountNanos,
		NumCopiesForSale:       numCopiesForSale,
		AvailableSerialNumbers: serialNumbersForSale,
	}
}
// These are to add to market on MINT
// AdminGetNFTDrop
func (fes *APIServer) GetMarketplaceRef(ww http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(io.LimitReader(req.Body, MaxRequestBodySizeBytes))
	requestData := AdminGetNFTDropRequest{}
	if err := decoder.Decode(&requestData); err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("GetMarketplaceRef: NFT Minted but adding to marketplace failed, contact Supernovas team for assistance. ERR1. 1"))
		return
	}

	var err error
	var dropEntryToReturn *NFTDropEntry
	if requestData.DropNumber < 0 {
		dropEntryToReturn, err = fes.GetLatestNFTDropEntry()
		if err != nil {
			_AddBadRequestError(ww, fmt.Sprintf("GetMarketplaceRef: NFT Minted but adding to marketplace failed, contact Supernovas team for assistance. ERR1. 2"))
			return
		}
	} else {
		// Look up the drop entry for the drop number given.
		dropEntryToReturn, err = fes.GetNFTDropEntry(uint64(requestData.DropNumber))
		if err != nil {
			_AddBadRequestError(ww, fmt.Sprintf(
				"GetMarketplaceRef: NFT Minted but adding to marketplace failed, contact Supernovas team for assistance. ERR1. 3"))
			return
		}
	}

	// Note that "dropEntryToReturn" can be nil if there are no entries in global state.
	var postEntryResponses []*PostEntryResponse
	if dropEntryToReturn != nil {
		postEntryResponses, err = fes.GetPostsForNFTDropEntry(dropEntryToReturn)
		if err != nil {
			_AddBadRequestError(ww, fmt.Sprintf("GetMarketplaceRef: NFT Minted but adding to marketplace failed, contact Supernovas team for assistance. ERR1. 4"))
			return
		}
	}

	// Return all the data associated with the transaction in the response
	res := AdminGetNFTDropResponse{
		DropEntry: dropEntryToReturn,
		Posts:     postEntryResponses,
	}

	if err = json.NewEncoder(ww).Encode(res); err != nil {
		_AddInternalServerError(ww, fmt.Sprintf("GetMarketplaceRef: NFT Minted but adding to marketplace failed, contact Supernovas team for assistance. ERR1. 5"))
		return
	}
}
// AdminUpdateNFTDrop
func (fes *APIServer) AddToMarketplace(ww http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(io.LimitReader(req.Body, MaxRequestBodySizeBytes))
	requestData := AdminUpdateNFTDropRequest{}
	err := decoder.Decode(&requestData)
	if err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("AddToMarketplace: NFT Minted but adding to marketplace failed, contact Supernovas team for assistance. ERR. 1"))
		return
	}

	if requestData.DropNumber < 1 {
		_AddBadRequestError(ww, fmt.Sprintf(
			"AddToMarketplace: NFT Minted but adding to marketplace failed, contact Supernovas team for assistance. ERR. 2"))
		return
	}

	if requestData.DropTstampNanos < 0 {
		_AddBadRequestError(ww, fmt.Sprintf(
			"AddToMarketplace: NFT Minted but adding to marketplace failed, contact Supernovas team for assistance. ERR. 3"))
		return
	}

	if requestData.NFTHashHexToAdd != "" && requestData.NFTHashHexToRemove != "" {
		_AddBadRequestError(ww, fmt.Sprint(
			"AddToMarketplace: NFT Minted but adding to marketplace failed, contact Supernovas team for assistance. ERR. 3"))
		return
	}

	var latestDropEntry *NFTDropEntry
	latestDropEntry, err = fes.GetLatestNFTDropEntry()
	if err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("AddToMarketplace: NFT Minted but adding to marketplace failed, contact Supernovas team for assistance. ERR. 4"))
		return
	}

	// Now for the business.
	var updatedDropEntry *NFTDropEntry
	currentTime := uint64(time.Now().UnixNano())
	if uint64(requestData.DropNumber) > latestDropEntry.DropNumber {
		// If we make it here, we are making a new drop. Run some checks to make sure that the
		// timestamp provided make sense.
		if latestDropEntry.DropTstampNanos > currentTime {
			_AddBadRequestError(ww, fmt.Sprint(
				"AddToMarketplace: NFT Minted but adding to marketplace failed, contact Supernovas team for assistance. ERR. 5"))
			return
		}
		if uint64(requestData.DropTstampNanos) < currentTime {
			_AddBadRequestError(ww, fmt.Sprint(
				"AddToMarketplace: NFT Minted but adding to marketplace failed, contact Supernovas team for assistance. ERR. 6"))
			return
		}
		if uint64(requestData.DropTstampNanos) < latestDropEntry.DropTstampNanos {
			_AddBadRequestError(ww, fmt.Sprint(
				"AddToMarketplace: NFT Minted but adding to marketplace failed, contact Supernovas team for assistance. ERR. 7"))
			return
		}

		// Regardless of the drop number provided, we force the new drop to be the previous number + 1.
		updatedDropEntry = &NFTDropEntry{
			DropNumber:      uint64(latestDropEntry.DropNumber + 1),
			DropTstampNanos: uint64(requestData.DropTstampNanos),
		}

	} else {
		// In this case, we are updating an existing drop.
		updatedDropEntry = latestDropEntry
		if uint64(requestData.DropNumber) != latestDropEntry.DropNumber {
			updatedDropEntry, err = fes.GetNFTDropEntry(uint64(requestData.DropNumber))
			if err != nil {
				_AddBadRequestError(ww, fmt.Sprintf(
					"AddToMarketplace: NFT Minted but adding to marketplace failed, contact Supernovas team for assistance. ERR. 8"))
				return
			}
		}

		// There are only two possible drops that can be updated (you can't update past drops):
		//   - The current "active" drop.
		//   - The next "pending" drop.
		canUpdateDrop := false
		latestDropIsPending := latestDropEntry.DropTstampNanos > currentTime
		if latestDropIsPending && uint64(requestData.DropNumber) >= latestDropEntry.DropNumber-1 {
			// In this case their is a pending drop so the latest drop and the previous drop are editable.
			canUpdateDrop = true
		} else if !latestDropIsPending && uint64(requestData.DropNumber) == latestDropEntry.DropNumber {
			// In this case there is no pending drop so you can only update the latest drop.
			canUpdateDrop = true
		}

		if !canUpdateDrop {
			_AddBadRequestError(ww, fmt.Sprintf(
				"AddToMarketplace: NFT Minted but adding to marketplace failed, contact Supernovas team for assistance. ERR. 9"))
			return
		}

		// Update IsActive.
		updatedDropEntry.IsActive = requestData.IsActive

		// Consider updating DropTstampNanos.
		if uint64(requestData.DropTstampNanos) > currentTime &&
			uint64(requestData.DropNumber) == latestDropEntry.DropNumber {
			updatedDropEntry.DropTstampNanos = uint64(requestData.DropTstampNanos)

		} else if uint64(requestData.DropTstampNanos) != updatedDropEntry.DropTstampNanos {
			_AddBadRequestError(ww, fmt.Sprintf(
				"AddToMarketplace: NFT Minted but adding to marketplace failed, contact Supernovas team for assistance. ERR. 10"))
			return
		}

		utxoView, err := fes.backendServer.GetMempool().GetAugmentedUniversalView()
		if err != nil {
			_AddBadRequestError(ww, fmt.Sprintf("AddToMarketplace: NFT Minted but adding to marketplace failed, contact Supernovas team for assistance. ERR. 11"))
			return
		}

		// Add new NFT hashes.
		if requestData.NFTHashHexToAdd != "" {
			// Decode the hash and make sure it is a valid NFT so that we can add it to the entry.
			postHash, err := GetPostHashFromPostHashHex(requestData.NFTHashHexToAdd)
			if err != nil {
				_AddBadRequestError(ww, fmt.Sprintf("AddToMarketplace: NFT Minted but adding to marketplace failed, contact Supernovas team for assistance. ERR. 12"))
				return
			}
			postEntry := utxoView.GetPostEntryForPostHash(postHash)
			if !postEntry.IsNFT {
				_AddBadRequestError(ww, fmt.Sprintf(
					"AddToMarketplace: NFT Minted but adding to marketplace failed, contact Supernovas team for assistance. ERR. 13"))
				return
			}

			updatedDropEntry.NFTHashes = append(updatedDropEntry.NFTHashes, postHash)
		}

	}

	// Set the updated drop entry.
	globalStateKey := GlobalStateKeyForNFTDropEntry(uint64(requestData.DropNumber))
	updatedDropEntryBuf := bytes.NewBuffer([]byte{})
	gob.NewEncoder(updatedDropEntryBuf).Encode(updatedDropEntry)
	err = fes.GlobalStatePut(globalStateKey, updatedDropEntryBuf.Bytes())
	if err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("AddToMarketplace: NFT Minted but adding to marketplace failed, contact Supernovas team for assistance. ERR. 14"))
		return
	}

	// Note that "dropEntryToReturn" can be nil if there are no entries in global state.
	postEntryResponses, err := fes.GetPostsForNFTDropEntry(updatedDropEntry)
	if err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("AddToMarketplace: NFT Minted but adding to marketplace failed, contact Supernovas team for assistance. ERR. 15"))
		return
	}

	// Return all the data associated with the transaction in the response
	res := AdminUpdateNFTDropResponse{
		DropEntry: updatedDropEntry,
		Posts:     postEntryResponses,
	}

	if err = json.NewEncoder(ww).Encode(res); err != nil {
		_AddInternalServerError(ww, fmt.Sprintf("AddToMarketplace: NFT Minted but adding to marketplace failed, contact Supernovas team for assistance. ERR. 16"))
		return
	}
}
type GetPostsStatelessRequest struct {
	// This is the PostHashHex of the post we want to start our paginated lookup at. We
	// will fetch up to "NumToFetch" posts after it, ordered by time stamp.  If no
	// PostHashHex is provided we will return the most recent posts.
	PostHashHex                string `safeForLogging:"true"`
	ReaderPublicKeyBase58Check string `safeForLogging:"true"`
	OrderBy                    string `safeForLogging:"true"`
	StartTstampSecs            uint64 `safeForLogging:"true"`
	PostContent                string `safeForLogging:"true"`
	NumToFetch                 int    `safeForLogging:"true"`

	// Note: if the GetPostsForFollowFeed option is passed, FetchSubcomments is currently ignored
	// (fetching comments / subcomments for the follow feed is currently unimplemented)
	FetchSubcomments bool `safeForLogging:"true"`

	// This gets posts by people that ReaderPublicKeyBase58Check follows.
	GetPostsForFollowFeed bool `safeForLogging:"true"`

	// This gets posts by people that ReaderPublicKeyBase58Check follows.
	GetPostsForGlobalWhitelist bool `safeForLogging:"true"`

	// This gets posts sorted by deso
	GetPostsByDESO  bool `safeForLogging:"true"`
	GetPostsByClout bool // Deprecated

	// This only gets posts that include media, like photos and videos
	MediaRequired bool `safeForLogging:"true"`

	PostsByDESOMinutesLookback uint64 `safeForLogging:"true"`

	// If set to true, then the posts in the response will contain a boolean about whether they're in the global feed
	AddGlobalFeedBool bool `safeForLogging:"true"`
}
/ GetPostsStateless ...
func (fes *APIServer) GetPostsStateless(ww http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(io.LimitReader(req.Body, MaxRequestBodySizeBytes))
	requestData := GetPostsStatelessRequest{}
	if err := decoder.Decode(&requestData); err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("GetPostsStateless: Problem parsing request body: %v", err))
		return
	}

	// Decode the reader public key into bytes. Default to nil if no pub key is passed in.
	var readerPublicKeyBytes []byte
	var err error
	if requestData.ReaderPublicKeyBase58Check != "" {

		readerPublicKeyBytes, _, err = lib.Base58CheckDecode(requestData.ReaderPublicKeyBase58Check)
		if err != nil {
			_AddBadRequestError(ww, fmt.Sprintf("GetPostsStateless: Problem decoding user public key: %v", err))
			return
		}
	}

	var startPostHash *lib.BlockHash
	if requestData.PostHashHex != "" {
		// Decode the postHash.  This will give us the location where we start our paginated search.
		startPostHash, err = GetPostHashFromPostHashHex(requestData.PostHashHex)
		if err != nil {
			_AddBadRequestError(ww, fmt.Sprintf("GetPostsStateless: %v", err))
			return
		}
	}

	// Default to 50 posts fetched.
	numToFetch := 50
	if requestData.NumToFetch != 0 {
		numToFetch = requestData.NumToFetch
	}

	if startPostHash == nil && numToFetch == 1 {
		_AddBadRequestError(ww, fmt.Sprintf("GetPostsStateless: Must provide PostHashHex when NumToFetch is 1"))
		return
	}

	// Get a view with all the mempool transactions (used to get all posts / reader state).
	utxoView, err := fes.backendServer.GetMempool().GetAugmentedUniversalView()
	if err != nil {a
		_AddBadRequestError(ww, fmt.Sprintf("GetPostsStateless: Error fetching mempool view"))
		return
	}

	// Get all the PostEntries
	var postEntries []*lib.PostEntry
	var commentsByPostHash map[lib.BlockHash][]*lib.PostEntry
	var profileEntryMap map[lib.PkMapKey]*lib.ProfileEntry
	var readerStateMap map[lib.BlockHash]*lib.PostEntryReaderState
	if requestData.GetPostsForFollowFeed {
		postEntries,
			profileEntryMap,
			readerStateMap,
			err = fes.GetPostEntriesForFollowFeed(startPostHash, readerPublicKeyBytes, numToFetch, utxoView, requestData.MediaRequired)
		// if we're getting posts for follow feed, no comments are returned (they aren't necessary)
		commentsByPostHash = make(map[lib.BlockHash][]*lib.PostEntry)
	} else if requestData.GetPostsForGlobalWhitelist {
		postEntries,
			profileEntryMap,
			readerStateMap,
			err = fes.GetPostEntriesForGlobalWhitelist(startPostHash, readerPublicKeyBytes, numToFetch, utxoView, requestData.MediaRequired)
		// if we're getting posts for the global whitelist, no comments are returned (they aren't necessary)
		commentsByPostHash = make(map[lib.BlockHash][]*lib.PostEntry)
	} else if requestData.GetPostsByDESO || requestData.GetPostsByClout {
		postEntries,
			profileEntryMap,
			err = fes.GetPostEntriesByDESOAfterTimePaginated(readerPublicKeyBytes, requestData.PostsByDESOMinutesLookback, numToFetch)
	} else {
		postEntries,
			commentsByPostHash,
			profileEntryMap,
			readerStateMap,
			err = fes.GetPostEntriesByTimePaginated(startPostHash, readerPublicKeyBytes, numToFetch, utxoView)
	}

	if err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("GetPostsStateless: Error fetching posts: %v", err))
		return
	}

	// Get a utxoView.
	utxoView, err = fes.backendServer.GetMempool().GetAugmentedUniversalView()
	if err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("GetPostsStateless: Error constucting utxoView: %v", err))
		return
	}

	blockedPubKeys, err := fes.GetBlockedPubKeysForUser(readerPublicKeyBytes)
	if err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("GetPostsStateless: Error fetching blocked pub keys for user: %v", err))
		return
	}

	postEntryResponses := []*PostEntryResponse{}
	for _, postEntry := range postEntries {
		// If the creator who posted postEntry is in the map of blocked pub keys, skip this postEntry
		if _, ok := blockedPubKeys[lib.PkToString(postEntry.PosterPublicKey, fes.Params)]; !ok {
			var postEntryResponse *PostEntryResponse
			postEntryResponse, err = fes._postEntryToResponse(postEntry, requestData.AddGlobalFeedBool, fes.Params, utxoView, readerPublicKeyBytes, 2)
			if err != nil {
				// Just ignore posts that fail to convert for whatever reason.
				continue
			}
			profileEntryFound := profileEntryMap[lib.MakePkMapKey(postEntry.PosterPublicKey)]
			postEntryResponse.ProfileEntryResponse = fes._profileEntryToResponse(
				profileEntryFound, utxoView)
			commentsFound := commentsByPostHash[*postEntry.PostHash]
			for _, commentEntry := range commentsFound {
				if _, ok = blockedPubKeys[lib.PkToString(commentEntry.PosterPublicKey, fes.Params)]; !ok {
					commentResponse, err := fes._getCommentResponse(commentEntry, profileEntryMap, requestData.AddGlobalFeedBool, utxoView, readerPublicKeyBytes)
					if fes._shouldSkipCommentResponse(commentResponse, err) {
						continue
					}

					// Fetch subcomments if needed
					if requestData.FetchSubcomments {
						subcommentsFound := commentsByPostHash[*commentEntry.PostHash]
						for _, subCommentEntry := range subcommentsFound {
							subcommentResponse, err := fes._getCommentResponse(subCommentEntry, profileEntryMap, requestData.AddGlobalFeedBool, utxoView, readerPublicKeyBytes)
							if fes._shouldSkipCommentResponse(subcommentResponse, err) {
								continue
							}
							commentResponse.Comments = append(commentResponse.Comments, subcommentResponse)
						}
						postEntryResponse.Comments = append(postEntryResponse.Comments, commentResponse)
					}
				}
			}
			postEntryResponse.PostEntryReaderState = readerStateMap[*postEntry.PostHash]
			postEntryResponses = append(postEntryResponses, postEntryResponse)
		}
	}

	if requestData.PostContent != "" {
		lowercaseFilter := strings.ToLower(requestData.PostContent)
		filteredResponses := []*PostEntryResponse{}
		for _, postRes := range postEntryResponses {
			if strings.Contains(strings.ToLower(postRes.Body), lowercaseFilter) {
				filteredResponses = append(filteredResponses, postRes)
			}
		}
		postEntryResponses = filteredResponses
	}

	if requestData.OrderBy == "newest" {
		// Now sort the post list on the timestamp
		sort.Slice(postEntryResponses, func(ii, jj int) bool {
			return postEntryResponses[ii].TimestampNanos > postEntryResponses[jj].TimestampNanos
		})
	} else if requestData.OrderBy == "oldest" {
		sort.Slice(postEntryResponses, func(ii, jj int) bool {
			return postEntryResponses[ii].TimestampNanos < postEntryResponses[jj].TimestampNanos
		})
	} else if requestData.OrderBy == "last_comment" {
		sort.Slice(postEntryResponses, func(ii, jj int) bool {
			lastCommentTimeii := uint64(0)
			if len(postEntryResponses[ii].Comments) > 0 {
				lastCommentTimeii = postEntryResponses[ii].Comments[len(postEntryResponses[ii].Comments)-1].TimestampNanos
			}
			lastCommentTimejj := uint64(0)
			if len(postEntryResponses[jj].Comments) > 0 {
				lastCommentTimejj = postEntryResponses[jj].Comments[len(postEntryResponses[jj].Comments)-1].TimestampNanos
			}
			return lastCommentTimeii > lastCommentTimejj
		})
	}

	// Return the posts found.
	res := &GetPostsStatelessResponse{
		PostsFound: postEntryResponses,
	}
	if err := json.NewEncoder(ww).Encode(res); err != nil {
		_AddBadRequestError(ww, fmt.Sprintf(
			"GetPostsStateless: Problem encoding response as JSON: %v", err))
		return
	}
}
type GetNFTShowcaseRequestPaginated struct {
	ReaderPublicKeyBase58Check string `safeForLogging:"true"`
	PostHashHex                string `safeForLogging:"true"`
	NumToFetch                 int    `safeForLogging:"true"`
}

func (fes *APIServer) GetNFTShowcasePaginated(ww http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(io.LimitReader(req.Body, MaxRequestBodySizeBytes))
	requestData := GetNFTShowcaseRequestPaginated{}
	if err := decoder.Decode(&requestData); err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("GetNFTShowcase: Error parsing request body: %v", err))
		return
	}

	var readerPublicKeyBytes []byte
	var err error
	if requestData.ReaderPublicKeyBase58Check != "" {
		readerPublicKeyBytes, _, err = lib.Base58CheckDecode(requestData.ReaderPublicKeyBase58Check)
		if err != nil {
			_AddBadRequestError(ww, fmt.Sprintf("GetNFTShowcase: Problem decoding reader public key: %v", err))
			return
		}
	}

	dropEntry, err := fes.GetLatestNFTDropEntry()
	if err != nil {
		_AddInternalServerError(ww, fmt.Sprintf("GetNFTShowcase: Problem getting latest drop: %v", err))
		return
	}

	currentTime := uint64(time.Now().UnixNano())
	if dropEntry.DropTstampNanos > currentTime {
		// In this case, we have found a pending drop. We must go back one drop in order to
		// get the current active drop.
		if dropEntry.DropNumber == 1 {
			// If the pending drop is drop #1, we need to return a blank dropEntry.
			dropEntry = &NFTDropEntry{}
		}

		if dropEntry.DropNumber > 1 {
			dropNumToFetch := dropEntry.DropNumber - 1
			dropEntry, err = fes.GetNFTDropEntry(dropNumToFetch)
			if err != nil {
				_AddInternalServerError(ww, fmt.Sprintf(
					"GetNFTShowcase: Problem getting drop #%d: %v", dropNumToFetch, err))
				return
			}
		}
	}
	// Addition
	// Check lastPostHash, if null start from the start
	var startPostHash *lib.BlockHash
	if requestData.PostHashHex != "" {
		// Decode the postHash.  This will give us the location where we start our paginated search.
		startPostHash, err = GetPostHashFromPostHashHex(requestData.PostHashHex)
		if err != nil {
			_AddBadRequestError(ww, fmt.Sprintf("GetPostsStateless: %v", err))
			return
		}
	}
	// Addition
	// Check request size, default to 50 posts fetched.
	numToFetch := 50
	if requestData.NumToFetch != 0 {
		numToFetch = requestData.NumToFetch
	}
	// Addition
	// filter collection only to have lastposthash -> next (request size)
	oneViewOfHashes []*lib.BlockHash
	if requestData.PostHashHex != "" {
		foundPost := false
		index := 0
		for _, nftHash := range dropEntry.NFTHashes {
			if index >= numToFetch {
				break
			}
			if foundPost == true {
				oneViewOfHashes = append(oneViewOfHashes, nftHash)
				index++
			}
			if nftHash == startPostHash {
				foundPost = true;
			}
		}
	} else {
		index := 0
		for _, nftHash := range dropEntry.NFTHashes {
			if index < numToFetch {
				oneViewOfHashes = append(oneViewOfHashes, nftHash)
				index++
			} else {
				break;
			}
		}
	}

	// Now that we have the drop entry, fetch the NFTs.
	utxoView, err := fes.backendServer.GetMempool().GetAugmentedUniversalView()
	if err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("GetNFTShowcase: Error getting utxoView: %v", err))
		return
	}

	var readerPKID *lib.PKID
	if requestData.ReaderPublicKeyBase58Check != "" {
		readerPKID = utxoView.GetPKIDForPublicKey(readerPublicKeyBytes).PKID
	}

	var nftCollectionResponses []*NFTCollectionResponsePlus

	// Addition
	for _, nftHash := range oneViewOfHashes {
		postEntry := utxoView.GetPostEntryForPostHash(nftHash)
		if postEntry == nil {
			_AddInternalServerError(ww, fmt.Sprint("GetNFTShowcase: Found nil post entry for NFT hash."))
			return
		}

		// Should fix the marketplace, stopped working once burn was implemented
		if postEntry.NumNFTCopiesBurned != postEntry.NumNFTCopies {
			nftKey := lib.MakeNFTKey(nftHash, 1)
			nftEntry := utxoView.GetNFTEntryForNFTKey(&nftKey)

			postEntryResponse, err := fes._postEntryToResponse(
				postEntry, false, fes.Params, utxoView, readerPublicKeyBytes, 2)
			if err != nil {
				_AddInternalServerError(ww, fmt.Sprint("GetNFTShowcase: Found invalid post entry for NFT hash."))
				return
			}
			postEntryResponse.PostEntryReaderState = utxoView.GetPostEntryReaderState(readerPublicKeyBytes, postEntry)
			nftCollectionResponsePlus := fes._nftEntryToNFTCollectionResponsePlus(nftEntry, postEntry.PosterPublicKey, postEntryResponse, utxoView, readerPKID)
			nftCollectionResponses = append(nftCollectionResponses, nftCollectionResponsePlus)
		}
	}

	// Return all the data associated with the transaction in the response
	res := GetNFTShowcaseResponsePlus{
		NFTCollections: nftCollectionResponses,
	}

	if err = json.NewEncoder(ww).Encode(res); err != nil {
		_AddInternalServerError(ww, fmt.Sprintf("GetNFTShowcase: Problem serializing object to JSON: %v", err))
		return
	}
}
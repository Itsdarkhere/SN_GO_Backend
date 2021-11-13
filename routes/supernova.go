package routes

import (	
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
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
// GetPostsForPublicKey... 
func (fes *APIServer) GetNFTsForPublicKeySupernovas(ww http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(io.LimitReader(req.Body, MaxRequestBodySizeBytes))
	requestData := GetPostsForPublicKeyRequest{}
	if err := decoder.Decode(&requestData); err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("GetPostsForPublicKey: Error parsing request body: %v", err))
		return
	}

	// Get a view
	utxoView, err := fes.backendServer.GetMempool().GetAugmentedUniversalView()
	if err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("GetPostsForPublicKey: Error getting utxoView: %v", err))
		return
	}

	// Decode the public key for which we are fetching posts. If a public key is not provided, use the username
	var publicKeyBytes []byte
	if requestData.PublicKeyBase58Check != "" {
		publicKeyBytes, _, err = lib.Base58CheckDecode(requestData.PublicKeyBase58Check)
		if err != nil {
			_AddBadRequestError(ww, fmt.Sprintf("GetPostsForPublicKey: Problem decoding user public key: %v", err))
			return
		}
	} else {
		username := requestData.Username
		profileEntry := utxoView.GetProfileEntryForUsername([]byte(username))

		// Return an error if we failed to find a profile entry
		if profileEntry == nil {
			_AddNotFoundError(ww, fmt.Sprintf("GetPostsForPublicKey: could not find profile for username: %v", username))
			return
		}
		publicKeyBytes = profileEntry.PublicKey
	}
	// Decode the reader's public key so we can fetch each post entry's reader state.
	var readerPk []byte
	if requestData.ReaderPublicKeyBase58Check != "" {
		readerPk, _, err = lib.Base58CheckDecode(requestData.ReaderPublicKeyBase58Check)
		if err != nil {
			_AddBadRequestError(ww, fmt.Sprintf("GetPostsForPublicKey: Problem decoding reader public key: %v", err))
			return
		}
	}

	var startPostHash *lib.BlockHash
	if requestData.LastPostHashHex != "" {
		// Get the StartPostHash from the LastPostHashHex
		startPostHash, err = GetPostHashFromPostHashHex(requestData.LastPostHashHex)
		if err != nil {
			_AddBadRequestError(ww, fmt.Sprintf("GetPostsForPublicKey: %v", err))
			return
		}
	}

	// Get Posts Ordered by time.
	posts, err := utxoView.GetPostsPaginatedForPublicKeyOrderedByTimestamp(publicKeyBytes, startPostHash, requestData.NumToFetch, requestData.MediaRequired)
	if err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("GetPostsForPublicKey: Problem getting paginated posts: %v", err))
		return
	}
	// Loop to get only NFTs: SUPERNOVAS ADDITION
	var PostEntries []*PostEntry
	for i := range posts {
		if posts[i].IsNFT {
			PostEntries = append(PostEntries, arr[i])
		}
	}
	// Make posts contain only the NFTS: SUPERNOVAS ADDITION
	posts = postEntries

	sort.Slice(posts, func(ii, jj int) bool {
		return posts[ii].TimestampNanos > posts[jj].TimestampNanos
	})

	// GetPostsPaginated returns all posts from the db and mempool, so we need to find the correct section of the
	// slice to return.
	if uint64(len(posts)) > requestData.NumToFetch || startPostHash != nil {
		startIndex := 0
		if startPostHash != nil {
			for ii, post := range posts {
				if reflect.DeepEqual(post.PostHash, startPostHash) {
					startIndex = ii + 1
					break
				}
			}
		}
		posts = posts[startIndex:lib.MinInt(len(posts), startIndex+int(requestData.NumToFetch))]
	}

	// Convert postEntries to postEntryResponses and fetch PostEntryReaderState for each post.
	var postEntryResponses []*PostEntryResponse
	for _, post := range posts {
		var postEntryResponse *PostEntryResponse
		postEntryResponse, err = fes._postEntryToResponse(post, true, fes.Params, utxoView, readerPk, 2)
		if err != nil {
			_AddBadRequestError(ww, fmt.Sprintf("GetPostsForPublicKey: Problem converting post entry to response: %v", err))
			return
		}
		if readerPk != nil {
			postEntryReaderState := utxoView.GetPostEntryReaderState(readerPk, post)
			postEntryResponse.PostEntryReaderState = postEntryReaderState
		}
		postEntryResponses = append(postEntryResponses, postEntryResponse)
	}
	// Return the last post hash hex in the slice to simplify pagination.
	var lastPostHashHex string
	if len(postEntryResponses) > 0 {
		lastPostHashHex = postEntryResponses[len(postEntryResponses)-1].PostHashHex
	}
	res := GetPostsForPublicKeyResponse{
		Posts:           postEntryResponses,
		LastPostHashHex: lastPostHashHex,
	}
	if err = json.NewEncoder(ww).Encode(res); err != nil {
		_AddInternalServerError(ww, fmt.Sprintf("GetPostsForPublicKey: Problem serializing object to JSON: %v", err))
		return
	}
}
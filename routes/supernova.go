package routes

import (	
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"	
	"context"
	"regexp"
	"encoding/base64"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/deso-protocol/core/lib"
)
const (
	host     = "65.108.105.40"
	port     = 65432
	user     = "supernovas-admin"
	password = "woebiuwecjlcasc283ryoih"
	dbname   = "supernovas-deso-db"
	categoryArt = "extra_data->>'category' = 'QXJ0' AND"
	categoryCollectibles = "extra_data->>'category' = 'Q29sbGVjdGlibGVz' AND"
	categoryGenerativeArt = "extra_data->>'category' = 'R2VuZXJhdGl2ZSBBcnQ=' AND"
	categoryMetaverseGaming = "extra_data->>'category' = 'TWV0YXZlcnNlICYgR2FtaW5n' AND"
	categoryMusic = "extra_data->>'category' = 'TXVzaWM=' AND"
	categoryProfilePicture = "extra_data->>'category' = 'UHJvZmlsZSBQaWN0dXJl' AND"
	categoryPhotography = "extra_data->>'category' = 'UGhvdG9ncmFwaHk=' AND"
	categoryImage = "extra_data->>'arweaveVideoSrc' IS NULL AND extra_data->>'arweaveAudioSrc' IS NULL AND"
	categoryVideo = "extra_data->>'arweaveVideoSrc' != '' AND"
	categoryAudio = "extra_data->>'arweaveAudioSrc' != '' AND"
	categoryFreshDrops = ""
	categoryCommunityFavourites = "true"
)

type PostResponse struct {
	Body string `db:"body"`
	ImageURLs []string
	VideoURLs []string
	PostHashHex string `db:"post_hash"`
    PosterPublicKeyBase58Check string
	ProfileEntryResponse *ProfileEntryResponse
    //parent_post_hash string `db:"parent_post_hash"`
    //reposted_post_hash string `db:"reposted_post_hash"`
    //quoted_repost bool `db:"quoted_repost"`
    TimestampNanos int64 `db:"timestamp"`
    IsHidden bool `db:"hidden"`
    LikeCount int64 `db:"like_count"`
    RepostCount int64 `db:"repost_count"`
    QuoteRepostCount int64 `db:"quote_repost_count"`
    DiamondCount int64 `db:"diamond_count"`
    CommentCount int64 `db:"comment_count"`
    IsPinned bool `db:"pinned"`
    IsNFT bool `db:"nft"`
    NumNFTCopies int64 `db:"num_nft_copies"`
    HasUnlockable bool `db:"unlockable"`
    NFTRoyaltyToCoinBasisPoints int64 `db:"creator_royalty_basis_points"`
    NFTRoyaltyToCreatorBasisPoints int64 `db:"coin_royalty_basis_points"`
    NumNFTCopiesForSale int64 `db:"num_nft_copies_for_sale"`
    NumNFTCopiesBurned int64 `db:"num_nft_copies_burned"`
	PostExtraData ExtraData `db:"extra_data"`
	// Information about the reader's state w/regard to this post (e.g. if they liked it).
	PostEntryReaderState *lib.PostEntryReaderState
}
// This is just to store values from the aggregate function that I dont need. 
type Waster struct {
	BidAmount int64 `db:"bid_amount"`
	LastAcceptedBidAmountNanos int64 `db:"last_accepted_bid_amount_nanos"`
	MinBidAmountNanos int64 `db:"min_bid_amount_nanos"`
}
type PostResponses struct {
	PostEntryResponse []*PostResponse
}
type PPKBytea struct {
	Poster_public_key []byte `db:"poster_public_key"`
}
type Body struct {
	Body string `db:"body"`
}
type BodyParts struct {
	ImageURLs []string
	VideoURLs []string
	Body string
}
type ExtraData map[string]string

func base64Encode(str string) string {
    return base64.StdEncoding.EncodeToString([]byte(str))
}

func base64Decode(str string) (string) {
    data, err := base64.StdEncoding.DecodeString(str)
    if err != nil {
        return ""
    }
    return string(data)
}

func JsonToStruct(data string) BodyParts {
	s := string(data)
	bp := BodyParts{}
	json.Unmarshal([]byte(s), &bp)
	return bp
}

// Connection pool
var pool *pgxpool.Pool

func CustomConnect() (*pgxpool.Pool, error) {
	// If we have a pool just return
	if pool != nil {
		return pool, nil
	}

	DATABASE_URL := "postgres://user_readonly:woebiuwecjlcasc283ryoih@65.108.105.40:65432/supernovas-deso-db"
	config, err := pgxpool.ParseConfig(DATABASE_URL)
	if err != nil {
		return nil, err
	}
	// Configs, Database dude said this is not needed
	//config.MaxConnIdleTime = 120 * time.Second
	//config.HealthCheckPeriod = 120 * time.Second
	//config.MaxConnIdleTime = 5 * time.Minute
	// setting pool
	pool, err = pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
        return nil, err
    }
	return pool, nil
}
// imx metadata connection pool
var poolETH *pgxpool.Pool

func CustomConnectETH() (*pgxpool.Pool, error) {
	// If we have a pool just return
	if poolETH != nil {
		return poolETH, nil
	}

	DATABASE_URL := "postgres://user_readonly:woebiuwecjlcasc283ryoih@65.108.105.40:65432/supernovas-data-db"
	config, err := pgxpool.ParseConfig(DATABASE_URL)
	if err != nil {
		return nil, err
	}
	// Configs, Database dude said this is not needed
	//config.MaxConnIdleTime = 120 * time.Second
	//config.HealthCheckPeriod = 120 * time.Second
	//config.MaxConnIdleTime = 5 * time.Minute
	// setting pool
	poolETH, err = pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
        return nil, err
    }
	return poolETH, nil
}
type GetCollectionInfoRequest struct {
	CollectionName string 
	CollectionCreatorName string 
}
type GetCollectionInfoResponse struct {
	Pieces int `db:"pieces"`
	OwnersAmount int `db:"owners_amount"`
	FloorPrice uint64 `db:"floor_price"`
	TradingVolume uint64 `db:"trading_volume"`
}
func (fes *APIServer) GetCollectionInfo(ww http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(io.LimitReader(req.Body, MaxRequestBodySizeBytes))
	requestData := GetCollectionInfoRequest{}
	if err := decoder.Decode(&requestData); err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("GetCollectionInfo: Error parsing request body: %v", err))
		return
	}
	// Confirm we have all needed fields
	if requestData.CollectionName == "" {
		_AddInternalServerError(ww, fmt.Sprintf("GetCollectionInfo: No CollectionName sent in request"))
		return
	}
	if requestData.CollectionCreatorName == "" {
		_AddInternalServerError(ww, fmt.Sprintf("GetCollectionInfo: No CollectionCreatorName sent in request"))
		return
	}

	// Get connection pool
	dbPool, err := CustomConnect()
	if err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("GetCollectionInfo: Error getting pool: %v", err))
		return
	}
	// get connection to pool
	conn, err := dbPool.Acquire(context.Background())
	if err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("GetCollectionInfo: Error cant connect to database: %v", err))
		conn.Release()
		return
	}

	// Release connection once function returns
	defer conn.Release();

	// Store response in this
	collectionInfoResponse := new(GetCollectionInfoResponse)

	// Query
	rows, err := conn.Query(context.Background(), 
	fmt.Sprintf(`SELECT COUNT(*) as pieces, COUNT(DISTINCT owner_pkid) as owners, 
	MIN(min_bid_amount_nanos) as floor_price,
	SUM(bid_amount_nanos) as trading_vol
	FROM pg_sn_collections
	INNER JOIN pg_posts ON pg_sn_collections.post_hash = pg_posts.post_hash
	INNER JOIN pg_nfts ON pg_sn_collections.post_hash = pg_nfts.nft_post_hash
	INNER JOIN pg_metadata_accept_nft_bids 
	ON pg_sn_collections.post_hash = pg_metadata_accept_nft_bids.nft_post_hash
	WHERE collection = '%v' AND creator_name = '%v'`, requestData.CollectionName, requestData.CollectionCreatorName))
	if err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("GetCollectionInfo: Error query failed: %v", err))
		return
	} else {

		// Defer closing rows
		defer rows.Close()

        // Next prepares the next row for reading.
        for rows.Next() {
			rows.Scan(&collectionInfoResponse.Pieces, &collectionInfoResponse.OwnersAmount, 
				&collectionInfoResponse.FloorPrice, &collectionInfoResponse.TradingVolume)
			// Check for errors
			if rows.Err() != nil {
				// if any error occurred while reading rows.
				_AddBadRequestError(ww, fmt.Sprintf("GetCollectionInfo: Error scanning to struct: %v", err))
				return
			}
        }
		// Send back response
		if err = json.NewEncoder(ww).Encode(collectionInfoResponse); err != nil {
			_AddInternalServerError(ww, fmt.Sprintf("GetCollectionInfo: Problem serializing object to JSON: %v", err))
			return
		}
		// Just to make sure call it here too, calling it multiple times has no side-effects
		conn.Release();
	}
}
type InsertIMXMetadataRequest {
	Name string `db:"name"`
	Description string `db:"description"`
	Image string `db:"image"`
	Image_url string `db:"image_url"`
}
type InsertIMXResponse struct {
	Response int 
}
func (fes *APIServer) InsertIMXMetadata(ww http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(io.LimitReader(req.Body, MaxRequestBodySizeBytes))
	requestData := InsertIMXMetadataRequest{}
	if err := decoder.Decode(&requestData); err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("InsertIMXMetadata: Error parsing request body: %v", err))
		return
	}
	// Confirm we have all needed fields
	if requestData.Name == "" {
		_AddInternalServerError(ww, fmt.Sprintf("InsertIMXMetadata: No Name sent in request"))
		return
	}
	name := requestData.Name
	if requestData.Description == "" {
		_AddInternalServerError(ww, fmt.Sprintf("InsertIMXMetadata: No Description sent in request"))
		return
	}
	description := requestData.Description
	if requestData.Image == "" {
		_AddInternalServerError(ww, fmt.Sprintf("InsertIMXMetadata: No Image sent in request"))
		return
	}
	image := requestData.Image
	if requestData.Image_url == "" {
		_AddInternalServerError(ww, fmt.Sprintf("InsertIMXMetadata: No Image_url sent in request"))
		return
	}
	image_url := requestData.Image_url

	// Get connection pool, NEW SINCE WE ARE USING ANOTHER DB THAN USUAL
	dbPool, err := CustomConnectETH()
	if err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("InsertIMXMetadata: Unable to connect to pool: %v", err))
		return
	}
	// get connection to pool
	conn, err := dbPool.Acquire(context.Background())
	if err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("InsertIMXMetadata: Unable to get db connection: %v", err))
		conn.Release()
		return
	}

	// Release connection once function returns
	defer conn.Release();

	id := 0
	err := conn.QueryRow(context.Background(), 
	fmt.Sprintf(
		`INSERT INTO pg_eth_metadata (name, description, image, image_url, token_id) 
		VALUES ('%v', 'v%', '%v', '%v', (SELECT MAX(token_id) + 1 FROM pg_eth_metadata)) RETURNING token_id`, 
		name, description, image, image_url)).Scan(&id)
	if err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("InsertIMXMetadata: Insert failed: %v", err))
		return
	}

	resp := InsertIMXResponse { 
		Response: id,
	}

	// Serialize response to JSON
	if err = json.NewEncoder(ww).Encode(resp); err != nil {
		_AddInternalServerError(ww, fmt.Sprintf("InsertIMXMetadata: Problem serializing object to JSON: %v", err))
		return
	}
	// Just to make sure call it here too, calling it multiple times has no side-effects
	conn.Release();

}
type GetIMXMetadataByIdResponse struct {
	Name string `db:"name"`
	Description string `db:"description"`
	Image string `db:"image"`
	Image_url string `db:"image_url"`
	Token_id int `db:"token_id"`
}
type getSingleIMXResponse struct {
	IMXMetadata: *GetIMXMetadataByIdResponse
}
func (fes *APIServer) GetIMXMetadataById(ww http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	id, idExists := vars["id"]
	if !idExists {
		_AddBadRequestError(ww, fmt.Sprintf("GetIMXMetadataById: Missing id"))
		return
	}

	// Get connection pool, NEW SINCE WE ARE USING ANOTHER DB THAN USUAL
	dbPool, err := CustomConnectETH()
	if err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("GetIMXMetadataById: Unable to connect to pool: %v", err))
		return
	}
	// get connection to pool
	conn, err := dbPool.Acquire(context.Background())
	if err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("GetIMXMetadataById: Unable to get db connection: %v", err))
		conn.Release()
		return
	}

	// Release connection once function returns
	defer conn.Release();

	// IMX response to store values in 
	singleIMXResponse := new(GetIMXMetadataByIdResponse)

	rows, err := conn.Query(context.Background(), 
	fmt.Sprintf("SELECT name, description, image, image_url, token_id FROM pg_eth_metadata WHERE token_id = '%v'", value))
	if err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("GetIMXMetadataById: Error in query: %v", err))
		return
	} else {
		defer rows.Close()

		for rows.Next() {
			rows.Scan(&singleIMXResponse.Name, &singleIMXResponse.Description, &singleIMXResponse.Image, &singleIMXResponse.Image_url, &singleIMXResponse.Token_id)
			if rows.Err() != nil {
				_AddBadRequestError(ww, fmt.Sprintf("GetIMXMetadataById: Error in scan: %v", err))
				return
			}
		}
		// Serialize response to JSON
		if err = json.NewEncoder(ww).Encode(singleIMXResponse); err != nil {
			_AddInternalServerError(ww, fmt.Sprintf("GetIMXMetadataById: Problem serializing object to JSON: %v", err))
			return
		}
		// Just to make sure call it here too, calling it multiple times has no side-effects
		conn.Release();
	}	



}
type GetUserCollectionsDataRequest struct {
	Username string
}
type GetUserCollectionsDataResponse struct {
	UserCollectionData []*UserCollectionsResponse
}
type UserCollectionsResponse struct {
	PostHashHex string `db:"post_hash"`
	Collection string `db:"collection"`
}
func (fes *APIServer) GetUserCollectionsData(ww http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(io.LimitReader(req.Body, MaxRequestBodySizeBytes))
	requestData := GetUserCollectionsDataRequest{}
	if err := decoder.Decode(&requestData); err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("GetUserCollectionsData: Error parsing request body: %v", err))
		return
	}
	if requestData.Username == "" {
		_AddInternalServerError(ww, fmt.Sprintf("GetUserCollectionsData: No Username sent in request"))
		return
	}

	// Get connection pool
	dbPool, err := CustomConnect()
	if err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("GetUserCollectionsData: Error getting pool: %v", err))
		return
	}
	// get connection to pool
	conn, err := dbPool.Acquire(context.Background())
	if err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("GetUserCollectionsData: Error cant connect to database: %v", err))
		conn.Release()
		return
	}

	// Release connection once function returns
	defer conn.Release();

	// Query
	rows, err := conn.Query(context.Background(), 
	fmt.Sprintf("SELECT encode(post_hash, 'hex') as post_hash, collection FROM pg_sn_collections WHERE creator_name = '%v'", requestData.Username))
	if err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("GetUserCollectionsData: Error query failed: %v", err))
		return
	} else {

		var collectionRows []*UserCollectionsResponse

		// Defer closing rows
		defer rows.Close()

        // Next prepares the next row for reading.
        for rows.Next() {
			// New post to insert values into
			collectionResponse := new(UserCollectionsResponse)
			rows.Scan(&collectionResponse.PostHashHex, &collectionResponse.Collection)
			// Check for errors
			if rows.Err() != nil {
				// if any error occurred while reading rows.
				_AddBadRequestError(ww, fmt.Sprintf("GetUserCollectionsData: Error scanning to struct: %v", err))
				return
			}
			// Append to array for returning
			collectionRows = append(collectionRows, collectionResponse)
        }

		resp := GetUserCollectionsDataResponse {
			UserCollectionData: collectionRows,
		}
		if err = json.NewEncoder(ww).Encode(resp); err != nil {
			_AddInternalServerError(ww, fmt.Sprintf("GetUserCollectionsData: Problem serializing object to JSON: %v", err))
			return
		}
		// Just to make sure call it here too, calling it multiple times has no side-effects
		conn.Release();
	}

}

type SortCollectionRequest struct {
	ReaderPublicKeyBase58Check string `safeForLogging:"true"`
	Username                   string `safeForLogging:"true"`
	CollectionName             string `safeForLogging:"true"`
	Offset 					   int64 `safeForLogging:"true"`
	Status 					   string `safeForLogging:"true"`
	Market                     string `safeForLogging:"true"`
	OrderByType				   string `safeForLogging:"true"`
}

type SortCollectionResponse struct {
	PostEntryResponse []*PostResponse
	CollectionName string
	CollectionDescription string
	CollectionBannerLocation string
}

func (fes *APIServer) SortCollection(ww http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(io.LimitReader(req.Body, MaxRequestBodySizeBytes))
	requestData := SortCollectionRequest{}
	if err := decoder.Decode(&requestData); err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("SortCollection: Error parsing request body: %v", err))
		return
	}
	if requestData.Username == "" {
		_AddInternalServerError(ww, fmt.Sprintf("SortCollection: No Username sent in request"))
		return
	}
	username := requestData.Username

	if requestData.CollectionName == "" {
		_AddInternalServerError(ww, fmt.Sprintf("SortCollection: No CollectionName sent in request"))
		return
	}
	collectionName := requestData.CollectionName

	// Get connection pool
	dbPool, err := CustomConnect()
	if err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("SortCollection: Error getting pool: %v", err))
		return
	}
	// get connection to pool
	conn, err := dbPool.Acquire(context.Background())
	if err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("SortCollection: Error cant connect to database: %v", err))
		conn.Release()
		return
	}

	// Release connection once function returns
	defer conn.Release();

	var readerPublicKeyBytes []byte
	var errr error
	if requestData.ReaderPublicKeyBase58Check != "" {
		readerPublicKeyBytes, _, errr = lib.Base58CheckDecode(requestData.ReaderPublicKeyBase58Check)
		if errr != nil {
			_AddBadRequestError(ww, fmt.Sprintf("SortCollection: Problem decoding reader public key: %v", errr))
			return
		}
	}

	// Cant and should not Inner join more than once on the same table
	// Change behaviour if two or more joins occur
	pg_nfts_inner_joined := false;

	has_bids_selected := false;

	sold_selected := false;

	var offset int64
	if requestData.Offset >= 0 {
		offset = requestData.Offset
	} else {
		offset = 0
	}

	// The basic variables are the base layer of the collections query
	// Based on user filtering we add options to it
	basic_select := `select encode(pg_posts.post_hash, 'hex') as post_hash, diamond_count, comment_count, 
	like_count, poster_public_key, body, timestamp, hidden, repost_count, quote_repost_count, 
	pinned, nft, num_nft_copies, unlockable, creator_royalty_basis_points,
	coin_royalty_basis_points, num_nft_copies_for_sale, num_nft_copies_burned, extra_data, collection, collection_description, banner_location`

	basic_from := ` FROM pg_sn_collections`

	basic_inner_join := " INNER JOIN pg_posts ON pg_sn_collections.post_hash = pg_posts.post_hash"

	basic_where := fmt.Sprintf(` WHERE nft = true AND num_nft_copies != num_nft_copies_burned AND creator_name = '%v' AND collection = '%v'`, username, collectionName)

	basic_group_by := " GROUP BY pg_posts.post_hash, collection, collection_description, banner_location"

	basic_offset := fmt.Sprintf(" OFFSET %v", offset)

	basic_order_by := " ORDER BY"

	basic_limit := ` LIMIT 30`

	// Switch for status 
	switch requestData.Status {
		case "all":
			// Add nothing
		case "for sale":
			basic_where = basic_where + " AND num_nft_copies_for_sale > 0"
		case "sold":
			basic_inner_join = basic_inner_join + " INNER JOIN pg_nfts ON pg_nfts.nft_post_hash = pg_posts.post_hash"
			basic_where = basic_where + " AND last_accepted_bid_amount_nanos > 0 AND num_nft_copies_for_sale = 0"
			// Change behaviour if someone tries joining twice
			pg_nfts_inner_joined = true;
			// Change it some more based on this
			sold_selected = true;
		// This used with an inner join to pg_nfts will not work 
		case "has bids":
			basic_inner_join = basic_inner_join + " INNER JOIN pg_nft_bids ON pg_nft_bids.nft_post_hash = pg_posts.post_hash"
			// Remove sold from the equasion
			basic_where = basic_where + " AND num_nft_copies_for_sale > 0"
			has_bids_selected = true;
		default:
			_AddBadRequestError(ww, "SortCollection: Error in status switch")
			return
	}

	// Switch for market 
	switch requestData.Market {
		case "all":
			// Do nothing
		case "primary": 
			if (pg_nfts_inner_joined) {
				basic_where = basic_where + " AND owner_pkid = poster_public_key"
			} else {
				basic_inner_join = basic_inner_join + " INNER JOIN pg_nfts ON pg_nfts.nft_post_hash = pg_posts.post_hash"
				basic_where = basic_where + " AND owner_pkid = poster_public_key"
				pg_nfts_inner_joined = true;
			}
		// High expense calculation
		case "secondary": 
			if (pg_nfts_inner_joined) {
				basic_where = basic_where + " AND owner_pkid != poster_public_key"
			} else {
				basic_inner_join = basic_inner_join + " INNER JOIN pg_nfts ON pg_nfts.nft_post_hash = pg_posts.post_hash"
				basic_where = basic_where + " AND owner_pkid != poster_public_key"
				pg_nfts_inner_joined = true;
			}
		default:
			_AddBadRequestError(ww, "SortCollection: Error in market switch")
			return
	}

	// Switch for sort
	switch requestData.OrderByType {
		case "most recent first":
			basic_order_by = basic_order_by + " timestamp desc"
		case "oldest first":
			basic_order_by = basic_order_by + " timestamp asc"
		case "highest price first":
			if (sold_selected) {
				// These MAX aggregate functions make sure the order by works basically
				// Aggregate must be used or alternative must add values from these to GROUP BY
				// But that results in duplicates
				basic_select = basic_select + ", MAX(pg_nfts.last_accepted_bid_amount_nanos) as last_accepted_bid_amount_nanos"
				basic_order_by = basic_order_by + " last_accepted_bid_amount_nanos desc"
			} else if (has_bids_selected) {
				basic_select = basic_select + ", MAX(pg_nft_bids.bid_amount_nanos) as bid_amount_nanos"
				basic_order_by = basic_order_by + " bid_amount_nanos desc"
			} else {
				if (pg_nfts_inner_joined) {
					basic_select = basic_select + ", MAX(pg_nfts.min_bid_amount_nanos) as min_bid_amount_nanos"
					basic_order_by = basic_order_by + " min_bid_amount_nanos desc"
				} else {
					basic_select = basic_select + ", MAX(pg_nfts.min_bid_amount_nanos) as min_bid_amount_nanos"
					basic_inner_join = basic_inner_join + " INNER JOIN pg_nfts ON pg_nfts.nft_post_hash = pg_posts.post_hash"
					basic_order_by = basic_order_by + " min_bid_amount_nanos desc"
					pg_nfts_inner_joined = true;
				}
			}
		case "lowest price first":
			if (sold_selected) {
				basic_select = basic_select + ", MAX(pg_nfts.last_accepted_bid_amount_nanos) as last_accepted_bid_amount_nanos"
				basic_order_by = basic_order_by + " last_accepted_bid_amount_nanos asc"
			} else if (has_bids_selected) {
				basic_select = basic_select + ", MAX(pg_nft_bids.bid_amount_nanos) as bid_amount_nanos"
				basic_order_by = basic_order_by + " bid_amount_nanos asc"
			} else {
				if (pg_nfts_inner_joined) {
					basic_select = basic_select + ", MAX(pg_nfts.min_bid_amount_nanos) as min_bid_amount_nanos"
					basic_order_by = basic_order_by + " min_bid_amount_nanos asc"
				} else {
					basic_select = basic_select + ", MAX(pg_nfts.min_bid_amount_nanos) as min_bid_amount_nanos"
					basic_inner_join = basic_inner_join + " INNER JOIN pg_nfts ON pg_nfts.nft_post_hash = pg_posts.post_hash"
					basic_order_by = basic_order_by + " min_bid_amount_nanos asc"
					pg_nfts_inner_joined = true;
				}
			}
		case "most likes first":
			basic_order_by = basic_order_by + " like_count desc"
		case "most diamonds first":
			basic_order_by = basic_order_by + " diamond_count desc"
		case "most comments first":
			basic_order_by = basic_order_by + " comment_count desc"
		case "most reposts first":
			basic_order_by = basic_order_by + " repost_count desc"
		default:
			_AddBadRequestError(ww, "SortCollection: Error in sort switch")
			return
	}

	// Concat the superstring 
	queryString := basic_select + basic_from + basic_inner_join + basic_where + basic_group_by + basic_order_by + basic_offset + basic_limit

	// For collection page we need these returned with the query
	collection := ""
	collection_description := ""
	banner_location := ""

	// Query
	rows, err := conn.Query(context.Background(), queryString)
	if err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("SortCollection: Error query failed: %v", err))
		return
	} else {

		var posts []*PostResponse

		// Defer closing rows
		defer rows.Close()

        // Next prepares the next row for reading.
        for rows.Next() {
			// New post to insert values into
			post := new(PostResponse)
			// Body is weird in db so I need this to parse it
			body := new(Body)
			// Need a holder var for the bytea format
			poster_public_key_bytea := new(PPKBytea)
            // Scan reads the values from the current row into tmp
			if (sold_selected && (requestData.OrderByType == "highest price first" || requestData.OrderByType == "lowest price first")) {
				wasteValue := new(Waster)
				rows.Scan(&post.PostHashHex, &post.DiamondCount, &post.CommentCount, &post.LikeCount,
					&poster_public_key_bytea.Poster_public_key, &body.Body, &post.TimestampNanos, &post.IsHidden, &post.RepostCount, 
					&post.QuoteRepostCount, &post.IsPinned, &post.IsNFT, &post.NumNFTCopies, &post.HasUnlockable,
					&post.NFTRoyaltyToCoinBasisPoints, &post.NFTRoyaltyToCreatorBasisPoints, &post.NumNFTCopiesForSale,
					&post.NumNFTCopiesBurned, &post.PostExtraData, &collection, &collection_description, &banner_location, &wasteValue.LastAcceptedBidAmountNanos)
				// Check for errors
				if rows.Err() != nil {
					// if any error occurred while reading rows.
					_AddBadRequestError(ww, fmt.Sprintf("SortCollection: Error scanning to struct: %v", err))
					return
				}

			} else if (has_bids_selected && (requestData.OrderByType == "highest price first" || requestData.OrderByType == "lowest price first")) {
				wasteValue := new(Waster)
				rows.Scan(&post.PostHashHex, &post.DiamondCount, &post.CommentCount, &post.LikeCount,
					&poster_public_key_bytea.Poster_public_key, &body.Body, &post.TimestampNanos, &post.IsHidden, &post.RepostCount, 
					&post.QuoteRepostCount, &post.IsPinned, &post.IsNFT, &post.NumNFTCopies, &post.HasUnlockable,
					&post.NFTRoyaltyToCoinBasisPoints, &post.NFTRoyaltyToCreatorBasisPoints, &post.NumNFTCopiesForSale,
					&post.NumNFTCopiesBurned, &post.PostExtraData, &collection, &collection_description, &banner_location, &wasteValue.BidAmount)
				// Check for errors
				if rows.Err() != nil {
					// if any error occurred while reading rows.
					_AddBadRequestError(ww, fmt.Sprintf("SortCollection: Error scanning to struct: %v", err))
					return
				}

			} else if (requestData.OrderByType == "highest price first" || requestData.OrderByType == "lowest price first") {
				wasteValue := new(Waster)
				rows.Scan(&post.PostHashHex, &post.DiamondCount, &post.CommentCount, &post.LikeCount,
					&poster_public_key_bytea.Poster_public_key, &body.Body, &post.TimestampNanos, &post.IsHidden, &post.RepostCount, 
					&post.QuoteRepostCount, &post.IsPinned, &post.IsNFT, &post.NumNFTCopies, &post.HasUnlockable,
					&post.NFTRoyaltyToCoinBasisPoints, &post.NFTRoyaltyToCreatorBasisPoints, &post.NumNFTCopiesForSale,
					&post.NumNFTCopiesBurned, &post.PostExtraData, &collection, &collection_description, &banner_location, &wasteValue.MinBidAmountNanos)
				// Check for errors
				if rows.Err() != nil {
					// if any error occurred while reading rows.
					_AddBadRequestError(ww, fmt.Sprintf("SortCollection: Error scanning to struct: %v", err))
					return
				}

			} else {
				rows.Scan(&post.PostHashHex, &post.DiamondCount, &post.CommentCount, &post.LikeCount,
					&poster_public_key_bytea.Poster_public_key, &body.Body, &post.TimestampNanos, &post.IsHidden, &post.RepostCount, 
					&post.QuoteRepostCount, &post.IsPinned, &post.IsNFT, &post.NumNFTCopies, &post.HasUnlockable,
					&post.NFTRoyaltyToCoinBasisPoints, &post.NFTRoyaltyToCreatorBasisPoints, &post.NumNFTCopiesForSale,
					&post.NumNFTCopiesBurned, &post.PostExtraData, &collection, &collection_description, &banner_location)
				// Check for errors
				if rows.Err() != nil {
					// if any error occurred while reading rows.
					_AddBadRequestError(ww, fmt.Sprintf("SortCollection: Error scanning to struct: %v", err))
					return
				}
			}
			if post.PostExtraData["name"] != "" {
				post.PostExtraData["name"] = base64Decode(post.PostExtraData["name"])
			}
			if post.PostExtraData["properties"] != "" {
				post.PostExtraData["properties"] = base64Decode(post.PostExtraData["properties"])
			}
			if post.PostExtraData["category"] != "" {
				post.PostExtraData["category"] = base64Decode(post.PostExtraData["category"])
			}
			if post.PostExtraData["Node"] != "" {
				post.PostExtraData["Node"] = base64Decode(post.PostExtraData["Node"])
			}
			if post.PostExtraData["arweaveVideoSrc"] != "" {
				post.PostExtraData["arweaveVideoSrc"] = base64Decode(post.PostExtraData["arweaveVideoSrc"])
			}
			if post.PostExtraData["arweaveAudioSrc"] != "" {
				post.PostExtraData["arweaveAudiooSrc"] = base64Decode(post.PostExtraData["arweaveAudioSrc"])
			}
			// Now break down the faulty body into a few parts
			content := JsonToStruct(body.Body)
			post.Body = content.Body
			post.ImageURLs = content.ImageURLs
			post.VideoURLs = content.VideoURLs

			// Get utxoView
			utxoView, err := fes.backendServer.GetMempool().GetAugmentedUniversalView()
			if err != nil {
				_AddBadRequestError(ww, fmt.Sprintf("SortCollection: Error getting utxoView: %v", err))
				return
			}

			// PKBytes to PKID
			var posterPKID *lib.PKID
			posterPKID = utxoView.GetPKIDForPublicKey(poster_public_key_bytea.Poster_public_key).PKID

			// PKID to profileEntry and PK
			profileEntry := utxoView.GetProfileEntryForPKID(posterPKID)
			var profileEntryResponse *ProfileEntryResponse
			var publicKeyBase58Check string
			if profileEntry != nil {
				profileEntryResponse = fes._profileEntryToResponse(profileEntry, utxoView)
				publicKeyBase58Check = profileEntryResponse.PublicKeyBase58Check
			} else {
				publicKey := utxoView.GetPublicKeyForPKID(posterPKID)
				publicKeyBase58Check = lib.PkToString(publicKey, fes.Params)
			}
			// Assign it to the post being returned
			post.PosterPublicKeyBase58Check = publicKeyBase58Check
			// Assign ProfileEntryResponse
			post.ProfileEntryResponse = profileEntryResponse
			// Decode the postHash.
			postHash, err := GetPostHashFromPostHashHex(post.PostHashHex)
			if err != nil {
				_AddBadRequestError(ww, fmt.Sprintf("SortCollection: %v", err))
				return
			}
			// Fetch the postEntry requested.
			postEntry := utxoView.GetPostEntryForPostHash(postHash)
			if postEntry == nil {
				_AddBadRequestError(ww, fmt.Sprintf("SortCollection: Could not find PostEntry for postHash"))
				return
			}
			// Get info regarding the readers interactions with the post
			post.PostEntryReaderState = utxoView.GetPostEntryReaderState(readerPublicKeyBytes, postEntry)
			// Append to array for returning
			posts = append(posts, post)
        }

		resp := SortCollectionResponse {
			PostEntryResponse: posts,
			CollectionName: collection,
			CollectionDescription: collection_description,
			CollectionBannerLocation: banner_location,
		}
		if err = json.NewEncoder(ww).Encode(resp); err != nil {
			_AddInternalServerError(ww, fmt.Sprintf("SortCollection: Problem serializing object to JSON: %v", err))
			return
		}
		// Just to make sure call it here too, calling it multiple times has no side-effects
		conn.Release();
	}

}
type CreateCollectionResponse struct {
	Response string 
}

type CreateCollectionRequest struct {
	PostHashHexArray        []string
	Username                string `safeForLogging:"true"`
	CollectionName          string `safeForLogging:"true"`
	CollectionDescription   string 
	CollectionBannerLocation string
}

func (fes *APIServer) CreateCollection(ww http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(io.LimitReader(req.Body, MaxRequestBodySizeBytes))
	requestData := CreateCollectionRequest{}
	if err := decoder.Decode(&requestData); err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("CreateCollection: Error parsing request body: %v", err))
		return
	}

	if len(requestData.PostHashHexArray) < 1 {
		_AddInternalServerError(ww, fmt.Sprintf("CreateCollection: No postHashHex sent in request"))
		return
	}
	hexArray := requestData.PostHashHexArray
	hexArrayPGFormat := "("
	// Loop through array and construct a string to use in insert
	for i := 0; i < len(hexArray); i++ {
		if (i + 1 == len(hexArray)) {
			hexArrayPGFormat = hexArrayPGFormat + "'" + hexArray[i] + "'"
		} else {
			hexArrayPGFormat = hexArrayPGFormat + "'" + hexArray[i] + "',"
		}
	}
	hexArrayPGFormat = hexArrayPGFormat + ")"
	if requestData.Username == "" {
		_AddInternalServerError(ww, fmt.Sprintf("CreateCollection: No Username sent in request"))
		return
	}
	username := requestData.Username
	if requestData.CollectionName == "" {
		_AddInternalServerError(ww, fmt.Sprintf("CreateCollection: No CollectionName sent in request"))
		return
	}
	collectionName := requestData.CollectionName

	if requestData.CollectionDescription == "" {
		_AddInternalServerError(ww, fmt.Sprintf("CreateCollection: No CollectionDescription sent in request"))
		return
	}
	collectionDescription := requestData.CollectionDescription

	if requestData.CollectionBannerLocation == "" {
		_AddInternalServerError(ww, fmt.Sprintf("CreateCollection: No CollectionBannerLocation sent in request"))
		return
	}
	collectionBannerLocation := requestData.CollectionBannerLocation

	conn, err := CustomConnect();
	if err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("CreateCollection: Error getting pool: %v", err))
		return
	}
	connection, err := conn.Acquire(context.Background())
	if err != nil {
		fmt.Println("CreateCollection: ERROR WITH POSTGRES CONNECTION")
	}
	
	selectString := fmt.Sprintf(`SELECT post_hash,'%v', '%v', '%v', '%v' FROM pg_posts 
	WHERE encode(post_hash, 'hex') IN %v`, username, collectionName, collectionDescription, collectionBannerLocation, hexArrayPGFormat)
	queryString := `INSERT INTO pg_sn_collections (post_hash, creator_name, collection, collection_description, banner_location) ` + selectString 

	// Query
	rows, err := connection.Query(context.Background(), queryString)
	if err != nil {
        _AddInternalServerError(ww, fmt.Sprintf("CreateCollection: Error connecting to postgres: ", err))
		return
    }

    for rows.Next() {
		s := ""
        if err := rows.Scan(&s); 
		err != nil {
            _AddInternalServerError(ww, fmt.Sprintf("CreateCollection: ERROR: ", err))
			return
        }
    }
    if err := rows.Err(); err != nil {
        _AddInternalServerError(ww, fmt.Sprintf("CreateCollection: ERROR: ", err))
		return
    }

	resp := CreateCollectionResponse { 
		Response: "Success",
	}

	if err = json.NewEncoder(ww).Encode(resp); err != nil {
		_AddInternalServerError(ww, fmt.Sprintf("CreateCollection: Problem serializing object to JSON: %v", err))
		return
	}

	// Just to make sure call it here too, calling it multiple times has no side-effects
	connection.Release();

}

type SortMarketplaceRequest struct {
	ReaderPublicKeyBase58Check string `safeForLogging:"true"`
	Offset int64 `safeForLogging:"true"`
	LowPrice uint64 `safeForLogging:"true"`
	HighPrice uint64 `safeForLogging:"true"`
	AuctionStatus string `safeForLogging:"true"`
	MarketType string `safeForLogging:"true"`
	Category string `safeForLogging:"true"`
	SortType string `safeForLogging:"true"`
	ContentFormat string `safeForLogging:"true"`
	CreatorsType string `safeForLogging:"true"`
}

func (fes *APIServer) SortMarketplace(ww http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(io.LimitReader(req.Body, MaxRequestBodySizeBytes))
	requestData := SortMarketplaceRequest{}
	if err := decoder.Decode(&requestData); err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("SortMarketplace: Error parsing request body: %v", err))
		return
	}
	// Get connection pool
	dbPool, err := CustomConnect()
	if err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("SortMarketplace: Error getting pool: %v", err))
		return
	}
	// get connection to pool
	conn, err := dbPool.Acquire(context.Background())
	if err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("SortMarketplace: Error cant connect to database: %v", err))
		conn.Release()
		return
	}

	// Release connection once function returns
	defer conn.Release();

	var readerPublicKeyBytes []byte
	var errr error
	if requestData.ReaderPublicKeyBase58Check != "" {
		readerPublicKeyBytes, _, errr = lib.Base58CheckDecode(requestData.ReaderPublicKeyBase58Check)
		if errr != nil {
			_AddBadRequestError(ww, fmt.Sprintf("GetNFTShowcase: Problem decoding reader public key: %v", errr))
			return
		}
	}

	// Cant and should not Inner join more than once on the same table
	// Change behaviour if two or more joins occur
	pg_nfts_inner_joined := false;

	has_bids_selected := false;

	sold_selected := false;

	var offset int64
	if requestData.Offset >= 0 {
		offset = requestData.Offset
	} else {
		offset = 0
	}

	// The basic variables are the base layer of the marketplace query
	// Based on user filtering we add options to it
	basic_select := `SELECT encode(post_hash, 'hex') as post_hash, diamond_count, comment_count, 
	like_count, 
	poster_public_key, body, timestamp, hidden, repost_count, quote_repost_count, 
	pinned, nft, num_nft_copies, unlockable, creator_royalty_basis_points,
	coin_royalty_basis_points, num_nft_copies_for_sale, num_nft_copies_burned, extra_data`

	basic_from := ` FROM pg_posts`

	basic_inner_join := ""

	basic_where := ` WHERE hidden = false AND nft = true 
	AND num_nft_copies != num_nft_copies_burned`

	basic_group_by := " GROUP BY post_hash"

	basic_offset := fmt.Sprintf(" OFFSET %v", offset)

	basic_limit := ` LIMIT 30`

	basic_order_by := " ORDER BY"

	// The switches below are in order, changing would cause problems

	// Switch for status 
	switch requestData.AuctionStatus {
		case "all":
			// Add nothing
		case "for sale":
			basic_where = basic_where + " AND num_nft_copies_for_sale > 0"
		case "sold":
			basic_inner_join = basic_inner_join + " INNER JOIN pg_nfts ON pg_nfts.nft_post_hash = post_hash"
			basic_where = basic_where + " AND last_accepted_bid_amount_nanos > 0 AND num_nft_copies_for_sale = 0"
			// Change behaviour if someone tries joining twice
			pg_nfts_inner_joined = true;
			// Change it some more based on this
			sold_selected = true;
		// This used with an inner join to pg_nfts will not work 
		case "has bids":
			basic_inner_join = basic_inner_join + " INNER JOIN pg_nft_bids ON pg_nft_bids.nft_post_hash = post_hash"
			// Remove sold from the equasion
			basic_where = basic_where + " AND num_nft_copies_for_sale > 0"
			has_bids_selected = true;
		default:
			_AddBadRequestError(ww, "SortMarketplace: Error in status switch")
			return
	}
	lowPrice := requestData.LowPrice
	highPrice := requestData.HighPrice
	// Check if prices are positive and if lowPrice is smaller than HighPrice
	if lowPrice < highPrice && lowPrice >= 0 {
		// This is true only if sold
		if (pg_nfts_inner_joined) {
			basic_where = basic_where + fmt.Sprintf(" AND last_accepted_bid_amount_nanos >= %v", lowPrice) + fmt.Sprintf(" AND last_accepted_bid_amount_nanos <= %v", highPrice)
		} else if (has_bids_selected) {
			basic_where = basic_where + fmt.Sprintf(" AND bid_amount_nanos >= %v", lowPrice) + fmt.Sprintf(" AND bid_amount_nanos <= %v", highPrice)
		} else {
			basic_inner_join = basic_inner_join + " INNER JOIN pg_nfts ON pg_nfts.nft_post_hash = post_hash"
			basic_where = basic_where + fmt.Sprintf(" AND min_bid_amount_nanos >= %v", lowPrice) + fmt.Sprintf(" AND min_bid_amount_nanos <= %v", highPrice)
			pg_nfts_inner_joined = true;
		}
	// If values equal each other do nothing, just to keep it simple for myself in the frontend.
	} else if lowPrice == highPrice {
		// Do nothing
	} else {
		_AddBadRequestError(ww, "SortMarketplace: Error in setting price range")
		return
	}
	
	// Switch for market 
	switch requestData.MarketType {
		case "all":
			// Do nothing
		case "primary": 
			if (pg_nfts_inner_joined) {
				basic_where = basic_where + " AND owner_pkid = poster_public_key"
			} else {
				basic_inner_join = basic_inner_join + " INNER JOIN pg_nfts ON pg_nfts.nft_post_hash = post_hash"
				basic_where = basic_where + " AND owner_pkid = poster_public_key"
				pg_nfts_inner_joined = true;
			}
		// High expense calculation 4.4s 
		case "secondary": 
			if (pg_nfts_inner_joined) {
				basic_where = basic_where + " AND owner_pkid != poster_public_key"
			} else {
				basic_inner_join = basic_inner_join + " INNER JOIN pg_nfts ON pg_nfts.nft_post_hash = post_hash"
				basic_where = basic_where + " AND owner_pkid != poster_public_key"
				pg_nfts_inner_joined = true;
			}
		default:
			_AddBadRequestError(ww, "SortMarketplace: Error in market switch")
			return
	}
	// switch for category
	switch requestData.Category {
		case "all":
			// Do nothing
		case "photography":
			basic_where = basic_where + " AND extra_data->>'category' = 'UGhvdG9ncmFwaHk='"
		case "profile picture":
			basic_where = basic_where + " AND extra_data->>'category' = 'UHJvZmlsZSBQaWN0dXJl'"
		case "music":
			basic_where = basic_where + " AND extra_data->>'category' = 'TXVzaWM='"
		case "metaverse":
			basic_where = basic_where + " AND extra_data->>'category' = 'TWV0YXZlcnNlICYgR2FtaW5n'"
		case "art":
			basic_where = basic_where + " AND extra_data->>'category' = 'QXJ0'"
		case "collectibles":
			basic_where = basic_where + " AND extra_data->>'category' = 'Q29sbGVjdGlibGVz'"
		case "generative art":
			basic_where = basic_where + " AND extra_data->>'category' = 'R2VuZXJhdGl2ZSBBcnQ='"
		default:
			_AddBadRequestError(ww, "SortMarketplace: Error in category switch")
			return
	}
	// Switch for sort
	switch requestData.SortType {
		case "most recent first":
			basic_order_by = basic_order_by + " timestamp desc"
		case "oldest first":
			basic_order_by = basic_order_by + " timestamp asc"
		case "highest price first":
			if (sold_selected) {
				// These MAX aggregate functions make sure the order by works basically
				// Aggregate must be used or alternative must add values from these to GROUP BY
				// But that results in duplicates
				basic_select = basic_select + ", MAX(pg_nfts.last_accepted_bid_amount_nanos) as last_accepted_bid_amount_nanos"
				basic_order_by = basic_order_by + " last_accepted_bid_amount_nanos desc"
			} else if (has_bids_selected) {
				basic_select = basic_select + ", MAX(pg_nft_bids.bid_amount_nanos) as bid_amount_nanos"
				basic_order_by = basic_order_by + " bid_amount_nanos desc"
			} else {
				if (pg_nfts_inner_joined) {
					basic_select = basic_select + ", MAX(pg_nfts.min_bid_amount_nanos) as min_bid_amount_nanos"
					basic_order_by = basic_order_by + " min_bid_amount_nanos desc"
				} else {
					basic_select = basic_select + ", MAX(pg_nfts.min_bid_amount_nanos) as min_bid_amount_nanos"
					basic_inner_join = basic_inner_join + " INNER JOIN pg_nfts ON pg_nfts.nft_post_hash = post_hash"
					basic_order_by = basic_order_by + " min_bid_amount_nanos desc"
					pg_nfts_inner_joined = true;
				}
			}
		case "lowest price first":
			if (sold_selected) {
				basic_select = basic_select + ", MAX(pg_nfts.last_accepted_bid_amount_nanos) as last_accepted_bid_amount_nanos"
				basic_order_by = basic_order_by + " last_accepted_bid_amount_nanos asc"
			} else if (has_bids_selected) {
				basic_select = basic_select + ", MAX(pg_nft_bids.bid_amount_nanos) as bid_amount_nanos"
				basic_order_by = basic_order_by + " bid_amount_nanos asc"
			} else {
				if (pg_nfts_inner_joined) {
					basic_select = basic_select + ", MAX(pg_nfts.min_bid_amount_nanos) as min_bid_amount_nanos"
					basic_order_by = basic_order_by + " min_bid_amount_nanos asc"
				} else {
					basic_select = basic_select + ", MAX(pg_nfts.min_bid_amount_nanos) as min_bid_amount_nanos"
					basic_inner_join = basic_inner_join + " INNER JOIN pg_nfts ON pg_nfts.nft_post_hash = post_hash"
					basic_order_by = basic_order_by + " min_bid_amount_nanos asc"
					pg_nfts_inner_joined = true;
				}
			}
		case "most likes first":
			basic_order_by = basic_order_by + " like_count desc"
		case "most diamonds first":
			basic_order_by = basic_order_by + " diamond_count desc"
		case "most comments first":
			basic_order_by = basic_order_by + " comment_count desc"
		case "most reposts first":
			basic_order_by = basic_order_by + " repost_count desc"
		default:
			_AddBadRequestError(ww, "SortMarketplace: Error in sort switch")
			return
	}
	// Switch for format 
	switch requestData.ContentFormat { // format_all = ""
		case "all":
		// Do nothing
		case "images":
			basic_where = basic_where + " AND body::json->>'ImageURLs' <> '[]' IS TRUE AND extra_data->>'arweaveAudioSrc' IS NULL"
		case "video":
			basic_where = basic_where + " AND (extra_data->>'arweaveVideoSrc' != '') OR (body::json->>'VideoURLs' != NULL)"
		case "music":
			basic_where = basic_where + " AND extra_data->>'arweaveAudioSrc' != ''"
		case "images video":
			basic_where = basic_where + " AND (body::json->>'ImageURLs' <> '[]' IS TRUE) OR (extra_data->>'arweaveVideoSrc' != '') OR (body::json->>'VideoURLs' != NULL)"
		case "images music":
			basic_where = basic_where + " AND (body::json->>'ImageURLs' <> '[]' IS TRUE)"
		case "music video":
			basic_where = basic_where + " AND extra_data->>'arweaveAudioSrc' != '' OR (extra_data->>'arweaveVideoSrc' != '') OR (body::json->>'VideoURLs' != NULL)"
		default:
			_AddBadRequestError(ww, "SortMarketplace: Error in format switch")
			return
	}
	// Switch for 
	switch requestData.CreatorsType {
		case "all":
			// Do nothing
		case "verified":
			basic_inner_join = basic_inner_join + " INNER JOIN pg_verified ON pg_verified.public_key = pg_posts.poster_public_key"
		default:
			_AddBadRequestError(ww, "SortMarketplace: Error in creators switch")
			return
	}

	// Concat the superstring 
	queryString := basic_select + basic_from + basic_inner_join + basic_where + basic_group_by + basic_order_by + basic_offset + basic_limit

	// Query
	rows, err := conn.Query(context.Background(), queryString)
	if err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("SortMarketplace: Error query failed: %v", err))
		return
	} else {

		var posts []*PostResponse

		// Defer closing rows
		defer rows.Close()

        // Next prepares the next row for reading.
        for rows.Next() {
			// New post to insert values into
			post := new(PostResponse)
			// Body is weird in db so I need this to parse it
			body := new(Body)
			// Need a holder var for the bytea format
			poster_public_key_bytea := new(PPKBytea)
            // Scan reads the values from the current row into tmp
			if (sold_selected && (requestData.SortType == "highest price first" || requestData.SortType == "lowest price first")) {
				wasteValue := new(Waster)
				rows.Scan(&post.PostHashHex, &post.DiamondCount, &post.CommentCount, &post.LikeCount,
					&poster_public_key_bytea.Poster_public_key, &body.Body, &post.TimestampNanos, &post.IsHidden, &post.RepostCount, 
					&post.QuoteRepostCount, &post.IsPinned, &post.IsNFT, &post.NumNFTCopies, &post.HasUnlockable,
					&post.NFTRoyaltyToCoinBasisPoints, &post.NFTRoyaltyToCreatorBasisPoints, &post.NumNFTCopiesForSale,
					&post.NumNFTCopiesBurned, &post.PostExtraData, &wasteValue.LastAcceptedBidAmountNanos)
				// Check for errors
				if rows.Err() != nil {
					// if any error occurred while reading rows.
					_AddBadRequestError(ww, fmt.Sprintf("SortMarketplace: Error scanning to struct: %v", err))
					return
				}

			} else if (has_bids_selected && (requestData.SortType == "highest price first" || requestData.SortType == "lowest price first")) {
				wasteValue := new(Waster)
				rows.Scan(&post.PostHashHex, &post.DiamondCount, &post.CommentCount, &post.LikeCount,
					&poster_public_key_bytea.Poster_public_key, &body.Body, &post.TimestampNanos, &post.IsHidden, &post.RepostCount, 
					&post.QuoteRepostCount, &post.IsPinned, &post.IsNFT, &post.NumNFTCopies, &post.HasUnlockable,
					&post.NFTRoyaltyToCoinBasisPoints, &post.NFTRoyaltyToCreatorBasisPoints, &post.NumNFTCopiesForSale,
					&post.NumNFTCopiesBurned, &post.PostExtraData, &wasteValue.BidAmount)
				// Check for errors
				if rows.Err() != nil {
					// if any error occurred while reading rows.
					_AddBadRequestError(ww, fmt.Sprintf("SortMarketplace: Error scanning to struct: %v", err))
					return
				}

			} else if (requestData.SortType == "highest price first" || requestData.SortType == "lowest price first") {
				wasteValue := new(Waster)
				rows.Scan(&post.PostHashHex, &post.DiamondCount, &post.CommentCount, &post.LikeCount,
					&poster_public_key_bytea.Poster_public_key, &body.Body, &post.TimestampNanos, &post.IsHidden, &post.RepostCount, 
					&post.QuoteRepostCount, &post.IsPinned, &post.IsNFT, &post.NumNFTCopies, &post.HasUnlockable,
					&post.NFTRoyaltyToCoinBasisPoints, &post.NFTRoyaltyToCreatorBasisPoints, &post.NumNFTCopiesForSale,
					&post.NumNFTCopiesBurned, &post.PostExtraData, &wasteValue.MinBidAmountNanos)
				// Check for errors
				if rows.Err() != nil {
					// if any error occurred while reading rows.
					_AddBadRequestError(ww, fmt.Sprintf("SortMarketplace: Error scanning to struct: %v", err))
					return
				}

			} else {
				rows.Scan(&post.PostHashHex, &post.DiamondCount, &post.CommentCount, &post.LikeCount,
					&poster_public_key_bytea.Poster_public_key, &body.Body, &post.TimestampNanos, &post.IsHidden, &post.RepostCount, 
					&post.QuoteRepostCount, &post.IsPinned, &post.IsNFT, &post.NumNFTCopies, &post.HasUnlockable,
					&post.NFTRoyaltyToCoinBasisPoints, &post.NFTRoyaltyToCreatorBasisPoints, &post.NumNFTCopiesForSale,
					&post.NumNFTCopiesBurned, &post.PostExtraData)
				// Check for errors
				if rows.Err() != nil {
					// if any error occurred while reading rows.
					_AddBadRequestError(ww, fmt.Sprintf("SortMarketplace: Error scanning to struct: %v", err))
					return
				}
			}
			if post.PostExtraData["name"] != "" {
				post.PostExtraData["name"] = base64Decode(post.PostExtraData["name"])
			}
			if post.PostExtraData["properties"] != "" {
				post.PostExtraData["properties"] = base64Decode(post.PostExtraData["properties"])
			}
			if post.PostExtraData["category"] != "" {
				post.PostExtraData["category"] = base64Decode(post.PostExtraData["category"])
			}
			if post.PostExtraData["Node"] != "" {
				post.PostExtraData["Node"] = base64Decode(post.PostExtraData["Node"])
			}
			if post.PostExtraData["arweaveVideoSrc"] != "" {
				post.PostExtraData["arweaveVideoSrc"] = base64Decode(post.PostExtraData["arweaveVideoSrc"])
			}
			if post.PostExtraData["arweaveAudioSrc"] != "" {
				post.PostExtraData["arweaveAudiooSrc"] = base64Decode(post.PostExtraData["arweaveAudioSrc"])
			}
			// Now break down the faulty body into a few parts
			content := JsonToStruct(body.Body)
			post.Body = content.Body
			post.ImageURLs = content.ImageURLs
			post.VideoURLs = content.VideoURLs

			// Get utxoView
			utxoView, err := fes.backendServer.GetMempool().GetAugmentedUniversalView()
			if err != nil {
				_AddBadRequestError(ww, fmt.Sprintf("SortMarketplace: Error getting utxoView: %v", err))
				return
			}

			// PKBytes to PKID
			var posterPKID *lib.PKID
			posterPKID = utxoView.GetPKIDForPublicKey(poster_public_key_bytea.Poster_public_key).PKID

			// PKID to profileEntry and PK
			profileEntry := utxoView.GetProfileEntryForPKID(posterPKID)
			var profileEntryResponse *ProfileEntryResponse
			var publicKeyBase58Check string
			if profileEntry != nil {
				profileEntryResponse = fes._profileEntryToResponse(profileEntry, utxoView)
				publicKeyBase58Check = profileEntryResponse.PublicKeyBase58Check
			} else {
				publicKey := utxoView.GetPublicKeyForPKID(posterPKID)
				publicKeyBase58Check = lib.PkToString(publicKey, fes.Params)
			}
			// Assign it to the post being returned
			post.PosterPublicKeyBase58Check = publicKeyBase58Check
			// Assign ProfileEntryResponse
			post.ProfileEntryResponse = profileEntryResponse
			// Decode the postHash.
			postHash, err := GetPostHashFromPostHashHex(post.PostHashHex)
			if err != nil {
				_AddBadRequestError(ww, fmt.Sprintf("SortMarketplace: %v", err))
				return
			}
			// Fetch the postEntry requested.
			postEntry := utxoView.GetPostEntryForPostHash(postHash)
			if postEntry == nil {
				_AddBadRequestError(ww, fmt.Sprintf("SortMarketplace: Could not find PostEntry for postHash"))
				return
			}
			// Get info regarding the readers interactions with the post
			post.PostEntryReaderState = utxoView.GetPostEntryReaderState(readerPublicKeyBytes, postEntry)
			// Append to array for returning
			posts = append(posts, post)
        }

		resp := PostResponses {
			PostEntryResponse: posts,
		}
		if err = json.NewEncoder(ww).Encode(resp); err != nil {
			_AddInternalServerError(ww, fmt.Sprintf("SortMarketplace: Problem serializing object to JSON: %v", err))
			return
		}
		// Just to make sure call it here too, calling it multiple times has no side-effects
		conn.Release();
	}

}
type SortCreatorsResponses struct {
	SortCreatorsResponses []*SortCreatorsProfileResponse
}
type SortCreatorsRequest struct {
	Offset int64 `safeForLogging:"true"`
	Verified string `safeForLogging:"true"`
}
type BodyJson struct {
	Body BodyData `db:"body"`
}
type BodyData map[string]interface{}

type SortCreatorsProfileResponse struct {
	Username string `db:"username"`
	ImageURLs []string `db:"body"`
}
func (fes *APIServer) SortCreators(ww http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(io.LimitReader(req.Body, MaxRequestBodySizeBytes))
	requestData := SortCreatorsRequest{}
	if err := decoder.Decode(&requestData); err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("SortCreators: Error parsing request body: %v", err))
		return
	}
	// Get connection pool
	dbPool, err := CustomConnect()
	if err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("SortCreators: Error getting pool: %v", err))
		return
	}
	// get connection to pool
	conn, err := dbPool.Acquire(context.Background())
	if err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("SortCreators: Error cant connect to database: %v", err))
		conn.Release()
		return
	}

	// Release connection once function returns
	defer conn.Release();

	var offset int64
	if requestData.Offset >= 0 {
		offset = requestData.Offset
	} else {
		offset = 0
	}

	// The basic variables are the base layer of the marketplace query
	// Based on user filtering we add options to it
	basic_select := `SELECT pg_profiles.username, CAST(body as JSON)`

	basic_from := ` FROM pg_profiles`

	basic_inner_join := " INNER JOIN pg_posts ON poster_public_key = public_key"

	basic_where := ` WHERE hidden = false AND nft = true AND body::json->>'ImageURLs' <> '[]' IS TRUE
	AND num_nft_copies != num_nft_copies_burned`

	basic_group_by := " GROUP BY pg_profiles.username, pg_posts.body"

	basic_offset := fmt.Sprintf(" OFFSET %v", offset)

	basic_limit := ` LIMIT 30`

	if requestData.Verified == "verified" {
		basic_inner_join = basic_inner_join + " INNER JOIN pg_verified ON pg_profiles.username = pg_verified.username"
	}

	// Concat the superstring 
	queryString := basic_select + basic_from + basic_inner_join + basic_where + basic_group_by + basic_offset + basic_limit

	// Query
	rows, err := conn.Query(context.Background(), queryString)
	if err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("SortCreators: Error query failed: %v", err))
		return
	} else {

		var profiles []*SortCreatorsProfileResponse

		// Defer closing rows
		defer rows.Close()

        // Next prepares the next row for reading.
        for rows.Next() {
			profile := new(SortCreatorsProfileResponse)
			body := new(Body)
            // Scan reads the values from the current row into tmp
            rows.Scan(&profile.Username, &body.Body)
				// Check for errors
				if rows.Err() != nil {
					// if any error occurred while reading rows.
					_AddBadRequestError(ww, fmt.Sprintf("SortCreators: Error scanning to struct: %v", err))
					return
				}
			// Now break down the faulty body into a few parts
			content := JsonToStruct(body.Body)
			profile.ImageURLs = content.ImageURLs
			profiles = append(profiles, profile)
        }

		resp := SortCreatorsResponses {
			SortCreatorsResponses: profiles,
		}
		if err = json.NewEncoder(ww).Encode(resp); err != nil {
			_AddInternalServerError(ww, fmt.Sprintf("SortCreators: Problem serializing object to JSON: %v", err))
			return
		}
		// Just to make sure call it here too, calling it multiple times has no side-effects
		conn.Release();
	}

}

func (fes *APIServer) GetCommunityFavourites(ww http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(io.LimitReader(req.Body, MaxRequestBodySizeBytes))
	requestData := GetNFTShowcaseRequest{}
	if err := decoder.Decode(&requestData); err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("GetCommunityFavourites: Error parsing request body: %v", err))
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

	// Get connection pool
	dbPool, err := CustomConnect()
	if err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("GetCommunityFavourites: Error getting pool: %v", err))
		return
	}
	// get connection to pool
	conn, err := dbPool.Acquire(context.Background())
	if err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("GetCommunityFavourites: Error cant connect to database: %v", err))
		conn.Release()
		return
	}

	// Release connection once function returns
	defer conn.Release();

	// Minus one week in Nanos
	timeUnix := uint64(time.Now().UnixNano()) - 259200000000000

	rows, err := conn.Query(context.Background(),
	fmt.Sprintf(`SELECT like_count, diamond_count, comment_count, encode(post_hash, 'hex') as post_hash, 
	poster_public_key, 
	body, timestamp, hidden, repost_count, quote_repost_count, 
	pinned, nft, num_nft_copies, unlockable, creator_royalty_basis_points,
	coin_royalty_basis_points, num_nft_copies_for_sale, num_nft_copies_burned, extra_data FROM pg_posts
	WHERE extra_data->>'Node' = 'OQ==' AND timestamp > %+v AND hidden = false AND nft = true 
	AND num_nft_copies != num_nft_copies_burned
	ORDER BY diamond_count desc, like_count desc, comment_count desc LIMIT 10`, timeUnix))
	if err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("GetCommunityFavourites: Error query failed: %v", err))
		return
	} else {

		var posts []*PostResponse

		// Defer closing rows
		defer rows.Close()

        // Next prepares the next row for reading.
        for rows.Next() {
			// New post to insert values into
			post := new(PostResponse)
			// Body is weird in db so I need this to parse it
			body := new(Body)
			// Need a holder var for the bytea format
			poster_public_key_bytea := new(PPKBytea)
            // Scan reads the values from the current row into tmp
            rows.Scan(&post.LikeCount, &post.DiamondCount, &post.CommentCount, &post.PostHashHex, 
				&poster_public_key_bytea.Poster_public_key, &body.Body, &post.TimestampNanos, &post.IsHidden, &post.RepostCount, 
				&post.QuoteRepostCount, &post.IsPinned, &post.IsNFT, &post.NumNFTCopies, &post.HasUnlockable,
				&post.NFTRoyaltyToCoinBasisPoints, &post.NFTRoyaltyToCreatorBasisPoints, &post.NumNFTCopiesForSale,
				&post.NumNFTCopiesBurned, &post.PostExtraData)
				// Check for errors
				if rows.Err() != nil {
					// if any error occurred while reading rows.
					_AddBadRequestError(ww, fmt.Sprintf("GetCommunityFavourites: Error scanning to struct: %v", err))
					return
				}
			if post.PostExtraData["name"] != "" {
				post.PostExtraData["name"] = base64Decode(post.PostExtraData["name"])
			}
			if post.PostExtraData["properties"] != "" {
				post.PostExtraData["properties"] = base64Decode(post.PostExtraData["properties"])
			}
			if post.PostExtraData["category"] != "" {
				post.PostExtraData["category"] = base64Decode(post.PostExtraData["category"])
			}
			if post.PostExtraData["Node"] != "" {
				post.PostExtraData["Node"] = base64Decode(post.PostExtraData["Node"])
			}
			if post.PostExtraData["arweaveVideoSrc"] != "" {
				post.PostExtraData["arweaveVideoSrc"] = base64Decode(post.PostExtraData["arweaveVideoSrc"])
			}
			if post.PostExtraData["arweaveAudioSrc"] != "" {
				post.PostExtraData["arweaveAudiooSrc"] = base64Decode(post.PostExtraData["arweaveAudioSrc"])
			}
			// Now break down the faulty body into a few parts
			content := JsonToStruct(body.Body)
			post.Body = content.Body
			post.ImageURLs = content.ImageURLs
			post.VideoURLs = content.VideoURLs

			// Get utxoView
			utxoView, err := fes.backendServer.GetMempool().GetAugmentedUniversalView()
			if err != nil {
				_AddBadRequestError(ww, fmt.Sprintf("GetCommunityFavourite: Error getting utxoView: %v", err))
				return
			}

			// PKBytes to PKID
			var posterPKID *lib.PKID
			posterPKID = utxoView.GetPKIDForPublicKey(poster_public_key_bytea.Poster_public_key).PKID

			// PKID to profileEntry and PK
			profileEntry := utxoView.GetProfileEntryForPKID(posterPKID)
			var profileEntryResponse *ProfileEntryResponse
			var publicKeyBase58Check string
			if profileEntry != nil {
				profileEntryResponse = fes._profileEntryToResponse(profileEntry, utxoView)
				publicKeyBase58Check = profileEntryResponse.PublicKeyBase58Check
			} else {
				publicKey := utxoView.GetPublicKeyForPKID(posterPKID)
				publicKeyBase58Check = lib.PkToString(publicKey, fes.Params)
			}
			// Assign it to the post being returned
			post.PosterPublicKeyBase58Check = publicKeyBase58Check
			// Assign ProfileEntryResponse
			post.ProfileEntryResponse = profileEntryResponse
			// Decode the postHash.
			postHash, err := GetPostHashFromPostHashHex(post.PostHashHex)
			if err != nil {
				_AddBadRequestError(ww, fmt.Sprintf("GetCommunityFavourites: %v", err))
				return
			}
			// Fetch the postEntry requested.
			postEntry := utxoView.GetPostEntryForPostHash(postHash)
			if postEntry == nil {
				_AddBadRequestError(ww, fmt.Sprintf("GetCommunityFavoutites: Could not fetch postEntry"))
				return
			}
			// Get info regarding the readers interactions with the post
			post.PostEntryReaderState = utxoView.GetPostEntryReaderState(readerPublicKeyBytes, postEntry)
			// Append to array for returning
			posts = append(posts, post)
        }

		resp := PostResponses {
			PostEntryResponse: posts,
		}
		if err = json.NewEncoder(ww).Encode(resp); err != nil {
			_AddInternalServerError(ww, fmt.Sprintf("GetCommunityFavourites: Problem serializing object to JSON: %v", err))
			return
		}
		// Just to make sure call it here too, calling it multiple times has no side-effects
		conn.Release();
	}
}
func (fes *APIServer) GetFreshDrops(ww http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(io.LimitReader(req.Body, MaxRequestBodySizeBytes))
	requestData := GetNFTShowcaseRequest{}
	if err := decoder.Decode(&requestData); err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("GetFreshDrops: Error parsing request body: %v", err))
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

	// Get connection pool
	dbPool, err := CustomConnect()
	if err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("GetFreshDrops: Error getting pool: %v", err))
		return
	}
	// get connection to pool
	conn, err := dbPool.Acquire(context.Background())
	if err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("GetCommunityFavourites: Error cant connect to database: %v", err))
		conn.Release()
		return
	}

	// Release connection once function returns
	defer conn.Release();


	rows, err := conn.Query(context.Background(),
	`SELECT like_count, diamond_count, comment_count, encode(post_hash, 'hex') as post_hash, 
	poster_public_key, 
	body, timestamp, hidden, repost_count, quote_repost_count, 
	pinned, nft, num_nft_copies, unlockable, creator_royalty_basis_points,
	coin_royalty_basis_points, num_nft_copies_for_sale, num_nft_copies_burned, extra_data FROM pg_posts
	WHERE extra_data->>'Node' = 'OQ==' AND hidden = false AND nft = true 
	AND num_nft_copies != num_nft_copies_burned
	ORDER BY timestamp desc LIMIT 8`)
	if err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("GetFreshDrops: Error query failed: %v", err))
		return
	} else {

		var posts []*PostResponse

		// Defer closing rows
		defer rows.Close()

        // Next prepares the next row for reading.
        for rows.Next() {
			// New post to insert values into
			post := new(PostResponse)
			// Body is weird in db so I need this to parse it
			body := new(Body)
			// Need a holder var for the bytea format
			poster_public_key_bytea := new(PPKBytea)
            // Scan reads the values from the current row into tmp
            rows.Scan(&post.LikeCount, &post.DiamondCount, &post.CommentCount, &post.PostHashHex, 
				&poster_public_key_bytea.Poster_public_key, &body.Body, &post.TimestampNanos, &post.IsHidden, &post.RepostCount, 
				&post.QuoteRepostCount, &post.IsPinned, &post.IsNFT, &post.NumNFTCopies, &post.HasUnlockable,
				&post.NFTRoyaltyToCoinBasisPoints, &post.NFTRoyaltyToCreatorBasisPoints, &post.NumNFTCopiesForSale,
				&post.NumNFTCopiesBurned, &post.PostExtraData)
				// Check for errors
				if rows.Err() != nil {
					// if any error occurred while reading rows.
					_AddBadRequestError(ww, fmt.Sprintf("GetFreshDrops: Error scanning to struct: %v", err))
					return
				}
			if post.PostExtraData["name"] != "" {
				post.PostExtraData["name"] = base64Decode(post.PostExtraData["name"])
			}
			if post.PostExtraData["properties"] != "" {
				post.PostExtraData["properties"] = base64Decode(post.PostExtraData["properties"])
			}
			if post.PostExtraData["category"] != "" {
				post.PostExtraData["category"] = base64Decode(post.PostExtraData["category"])
			}
			if post.PostExtraData["Node"] != "" {
				post.PostExtraData["Node"] = base64Decode(post.PostExtraData["Node"])
			}
			if post.PostExtraData["arweaveVideoSrc"] != "" {
				post.PostExtraData["arweaveVideoSrc"] = base64Decode(post.PostExtraData["arweaveVideoSrc"])
			}
			if post.PostExtraData["arweaveAudioSrc"] != "" {
				post.PostExtraData["arweaveAudiooSrc"] = base64Decode(post.PostExtraData["arweaveAudioSrc"])
			}
			// Now break down the faulty body into a few parts
			content := JsonToStruct(body.Body)
			post.Body = content.Body
			post.ImageURLs = content.ImageURLs
			post.VideoURLs = content.VideoURLs

			// Get utxoView
			utxoView, err := fes.backendServer.GetMempool().GetAugmentedUniversalView()
			if err != nil {
				_AddBadRequestError(ww, fmt.Sprintf("GetFreshDrops: Error getting utxoView: %v", err))
				return
			}

			// PKBytes to PKID
			var posterPKID *lib.PKID
			posterPKID = utxoView.GetPKIDForPublicKey(poster_public_key_bytea.Poster_public_key).PKID

			// PKID to profileEntry and PK
			profileEntry := utxoView.GetProfileEntryForPKID(posterPKID)
			var profileEntryResponse *ProfileEntryResponse
			var publicKeyBase58Check string
			if profileEntry != nil {
				profileEntryResponse = fes._profileEntryToResponse(profileEntry, utxoView)
				publicKeyBase58Check = profileEntryResponse.PublicKeyBase58Check
			} else {
				publicKey := utxoView.GetPublicKeyForPKID(posterPKID)
				publicKeyBase58Check = lib.PkToString(publicKey, fes.Params)
			}
			// Assign it to the post being returned
			post.PosterPublicKeyBase58Check = publicKeyBase58Check
			// Assign ProfileEntryResponse
			post.ProfileEntryResponse = profileEntryResponse
			// Decode the postHash.
			postHash, err := GetPostHashFromPostHashHex(post.PostHashHex)
			if err != nil {
				_AddBadRequestError(ww, fmt.Sprintf("GetCommunityFavourites: %v", err))
				return
			}
			// Fetch the postEntry requested.
			postEntry := utxoView.GetPostEntryForPostHash(postHash)
			if postEntry == nil {
				_AddBadRequestError(ww, fmt.Sprintf("GetFreshdrops: Could not fetch postEntry"))
				return
			}
			// Get info regarding the readers interactions with the post
			post.PostEntryReaderState = utxoView.GetPostEntryReaderState(readerPublicKeyBytes, postEntry)
			// Append to array for returning
			posts = append(posts, post)
        }

		resp := PostResponses {
			PostEntryResponse: posts,
		}
		if err = json.NewEncoder(ww).Encode(resp); err != nil {
			_AddInternalServerError(ww, fmt.Sprintf("GetFreshDrops: Problem serializing object to JSON: %v", err))
			return
		}
		// Just to make sure
		conn.Release()
	}
}
type GetNFTsByCategoryRequest struct {
	ReaderPublicKeyBase58Check string `safeForLogging:"true"`
	Category string `safeForLogging:"true"`
	Offset int64 `safeForLogging:"true"`
}
func (fes *APIServer) GetNFTsByCategory(ww http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(io.LimitReader(req.Body, MaxRequestBodySizeBytes))
	requestData := GetNFTsByCategoryRequest{}
	if err := decoder.Decode(&requestData); err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("GetNFTsByCategory: Error parsing request body: %v", err))
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

	categoryString := requestData.Category

	switch categoryString {
		case "art":
			categoryString = categoryArt
		case "collectibles":
			categoryString = categoryCollectibles
		case "generative":
			categoryString = categoryGenerativeArt
		case "metaverse":
			categoryString = categoryMetaverseGaming
		case "music":
			categoryString = categoryMusic
		case "profilepic":
			categoryString = categoryProfilePicture
		case "photography":
			categoryString = categoryPhotography
		case "fresh":
			categoryString = categoryFreshDrops
		case "communityfavourites":
			categoryString = categoryCommunityFavourites
		case "image":
			categoryString = categoryImage
		case "video":
			categoryString = categoryVideo
		case "audio":
			categoryString = categoryAudio
		default:
			_AddBadRequestError(ww, "GetNFTsByCategory: Error invalid category type")
			return
	}

	var offset int64
	if requestData.Offset >= 0 {
		offset = requestData.Offset
	} else {
		offset = 0
	}

	// Get connection pool
	dbPool, err := CustomConnect()
	if err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("GetCommunityFavourites: Error getting pool: %v", err))
		return
	}
	// get connection to pool
	conn, err := dbPool.Acquire(context.Background())
	if err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("GetCommunityFavourites: Error cant connect to database: %v", err))
		conn.Release()
		return
	}

	// Release connection once function returns
	defer conn.Release();

	// Combining this and the lower one def is something to do
	var queryString string
	// IF categoryString is true Order based on diamond count, this is only in communityFavourites
	if categoryString == "true" {
		queryString = `SELECT like_count, diamond_count, comment_count, encode(post_hash, 'hex') as post_hash, 
		poster_public_key, 
		body, timestamp, hidden, repost_count, quote_repost_count, 
		pinned, nft, num_nft_copies, unlockable, creator_royalty_basis_points,
		coin_royalty_basis_points, num_nft_copies_for_sale, num_nft_copies_burned, extra_data FROM pg_posts
		WHERE extra_data->>'Node' = 'OQ==' AND hidden = false AND nft = true 
		AND num_nft_copies != num_nft_copies_burned
		ORDER BY diamond_count desc, like_count desc, comment_count desc`
	} else {
		queryString = fmt.Sprintf(`SELECT like_count, diamond_count, comment_count, encode(post_hash, 'hex') as post_hash, 
		poster_public_key, 
		body, timestamp, hidden, repost_count, quote_repost_count, 
		pinned, nft, num_nft_copies, unlockable, creator_royalty_basis_points,
		coin_royalty_basis_points, num_nft_copies_for_sale, num_nft_copies_burned, extra_data FROM pg_posts
		WHERE extra_data->>'Node' = 'OQ==' AND %+v hidden = false AND nft = true 
		AND num_nft_copies != num_nft_copies_burned
		ORDER BY timestamp desc`, categoryString)
	}
	// So this
	queryString = queryString + fmt.Sprintf(" OFFSET %+v LIMIT 30", offset)

	rows, err := conn.Query(context.Background(), queryString)
	if err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("GetNFTsByCategory: Error query failed: %v", err))
		return
	} else {
		// carefully deferring Queries closing

		var posts []*PostResponse

		// Defer closing rows
		defer rows.Close()

        // Next prepares the next row for reading.
        for rows.Next() {
			// New post to insert values into
			post := new(PostResponse)
			// Body is weird in db so I need this to parse it
			body := new(Body)
			// Need a holder var for the bytea format
			poster_public_key_bytea := new(PPKBytea)
            // Scan reads the values from the current row into tmp
            rows.Scan(&post.LikeCount, &post.DiamondCount, &post.CommentCount, &post.PostHashHex, 
				&poster_public_key_bytea.Poster_public_key, &body.Body, &post.TimestampNanos, &post.IsHidden, &post.RepostCount, 
				&post.QuoteRepostCount, &post.IsPinned, &post.IsNFT, &post.NumNFTCopies, &post.HasUnlockable,
				&post.NFTRoyaltyToCoinBasisPoints, &post.NFTRoyaltyToCreatorBasisPoints, &post.NumNFTCopiesForSale,
				&post.NumNFTCopiesBurned, &post.PostExtraData)
				// Check for errors
				if rows.Err() != nil {
					// if any error occurred while reading rows.
					_AddBadRequestError(ww, fmt.Sprintf("GetNFTsByCategory: Error scanning to struct: %v", err))
					return
				}
			if post.PostExtraData["name"] != "" {
				post.PostExtraData["name"] = base64Decode(post.PostExtraData["name"])
			}
			if post.PostExtraData["properties"] != "" {
				post.PostExtraData["properties"] = base64Decode(post.PostExtraData["properties"])
			}
			if post.PostExtraData["category"] != "" {
				post.PostExtraData["category"] = base64Decode(post.PostExtraData["category"])
			}
			if post.PostExtraData["Node"] != "" {
				post.PostExtraData["Node"] = base64Decode(post.PostExtraData["Node"])
			}
			if post.PostExtraData["arweaveVideoSrc"] != "" {
				post.PostExtraData["arweaveVideoSrc"] = base64Decode(post.PostExtraData["arweaveVideoSrc"])
			}
			if post.PostExtraData["arweaveAudioSrc"] != "" {
				post.PostExtraData["arweaveAudiooSrc"] = base64Decode(post.PostExtraData["arweaveAudioSrc"])
			}
			// Now break down the faulty body into a few parts
			content := JsonToStruct(body.Body)
			post.Body = content.Body
			post.ImageURLs = content.ImageURLs
			post.VideoURLs = content.VideoURLs

			// Get utxoView
			utxoView, err := fes.backendServer.GetMempool().GetAugmentedUniversalView()
			if err != nil {
				_AddBadRequestError(ww, fmt.Sprintf("GetNFTsByCategory: Error getting utxoView: %v", err))
				return
			}

			// PKBytes to PKID
			var posterPKID *lib.PKID
			posterPKID = utxoView.GetPKIDForPublicKey(poster_public_key_bytea.Poster_public_key).PKID

			// PKID to profileEntry and PK
			profileEntry := utxoView.GetProfileEntryForPKID(posterPKID)
			var profileEntryResponse *ProfileEntryResponse
			var publicKeyBase58Check string
			if profileEntry != nil {
				profileEntryResponse = fes._profileEntryToResponse(profileEntry, utxoView)
				publicKeyBase58Check = profileEntryResponse.PublicKeyBase58Check
			} else {
				publicKey := utxoView.GetPublicKeyForPKID(posterPKID)
				publicKeyBase58Check = lib.PkToString(publicKey, fes.Params)
			}
			// Assign it to the post being returned
			post.PosterPublicKeyBase58Check = publicKeyBase58Check
			// Assign ProfileEntryResponse
			post.ProfileEntryResponse = profileEntryResponse
			// Decode the postHash.
			postHash, err := GetPostHashFromPostHashHex(post.PostHashHex)
			if err != nil {
				_AddBadRequestError(ww, fmt.Sprintf("GetCommunityFavourites: %v", err))
				return
			}
			// Fetch the postEntry requested.
			postEntry := utxoView.GetPostEntryForPostHash(postHash)
			if postEntry == nil {
				_AddBadRequestError(ww, fmt.Sprintf("GetCommunityFavourites: Could not fetch postEntry"))
				return
			}
			// Get info regarding the readers interactions with the post
			post.PostEntryReaderState = utxoView.GetPostEntryReaderState(readerPublicKeyBytes, postEntry)
			// Append to array for returning
			posts = append(posts, post)
        }

		resp := PostResponses {
			PostEntryResponse: posts,
		}
		if err = json.NewEncoder(ww).Encode(resp); err != nil {
			_AddInternalServerError(ww, fmt.Sprintf("GetNFTsByCategory: Problem serializing object to JSON: %v", err))
			return
		}
		// Just to make sure
		conn.Release()
	}
}
type NFTCollectionResponsePlus struct {
	NFTEntryResponse       *NFTEntryResponse     `json:",omitempty"`
	ProfileEntryResponse   *ProfileEntryResponse `json:",omitempty"`
	PostEntryResponse      *PostEntryResponse    `json:",omitempty"`
	HighestBidAmountNanos  uint64                `safeForLogging:"true"`
	LowestBidAmountNanos   uint64                `safeForLogging:"true"`
	HighestBuyNowPriceNanos *uint64               `safeForLogging:"true"`
	LowestBuyNowPriceNanos  *uint64               `safeForLogging:"true"`
	NumCopiesForSale       uint64                `safeForLogging:"true"`
	NumCopiesBuyNow         uint64                `safeForLogging:"true"`
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
	var numCopiesBuyNow uint64
	var highBuyNowPriceNanos *uint64
	var lowBuyNowPriceNanos *uint64
	serialNumbersForSale := []uint64{}
	for ii := uint64(1); ii <= postEntryResponse.NumNFTCopies; ii++ {
		nftKey := lib.MakeNFTKey(nftEntry.NFTPostHash, ii)
		nftEntryii := utxoView.GetNFTEntryForNFTKey(&nftKey)
		if nftEntryii != nil && nftEntryii.IsForSale {
			if nftEntryii.OwnerPKID != readerPKID {
				serialNumbersForSale = append(serialNumbersForSale, ii)
				if nftEntryii.IsBuyNow {
					if highBuyNowPriceNanos == nil || nftEntryii.BuyNowPriceNanos > *highBuyNowPriceNanos {
						highBuyNowPriceNanos = &nftEntryii.BuyNowPriceNanos
					}
					if lowBuyNowPriceNanos == nil || nftEntryii.BuyNowPriceNanos < *lowBuyNowPriceNanos {
						lowBuyNowPriceNanos = &nftEntryii.BuyNowPriceNanos
					}
				}
			}
			if nftEntryii.IsBuyNow {
				numCopiesBuyNow++
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
		HighestBuyNowPriceNanos: highBuyNowPriceNanos,
		LowestBuyNowPriceNanos:  lowBuyNowPriceNanos,
		NumCopiesForSale:       numCopiesForSale,
		NumCopiesBuyNow:         numCopiesBuyNow,
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
	err = fes.GlobalState.Put(globalStateKey, updatedDropEntryBuf.Bytes())
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
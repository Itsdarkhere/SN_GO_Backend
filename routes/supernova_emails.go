package routes

import (
	"context"
	"encoding/json"
	"net/http"
	"io"
	"fmt"
	sendinblue "github.com/sendinblue/APIv3-go-library/lib"
)
// This File is used to send transactional emails to Supernovas users
// These routes could very well be combined into one but...
// Id rather have it super clear what everything does on the frontend
// Rather than start sending wrong emails
// Enjoy, or dont <3

// Used for all routes as the success response
type EmailSuccessResponse struct {
	Success bool `json:"success"`
}

// Verify Email Template
type VerifyEmailEmailRequest struct {
	Username string `json:"username"`
	Link string `json:"link"`
	Email string `json:"email"`
}
// Sent to users after email has been verified ? Something along those lines
func (fes *APIServer) SendVerifyEmailEmail(ww http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(io.LimitReader(req.Body, MaxRequestBodySizeBytes))
	// Check Request 
	requestData := VerifyEmailEmailRequest{}
	if err := decoder.Decode(&requestData); err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("SendVerifyEmailEmail: Error parsing request body: %v", err))
		return
	}
	// Get context
	var ctx context.Context

	// Sendingblue configuration
	cfg := sendinblue.NewConfiguration()
	// Api key
	cfg.AddDefaultHeader("api-key", "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")
	// Partner key
	cfg.AddDefaultHeader("partner-key", "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")

	// Connect to SendinBlue Api client
	sib := sendinblue.NewAPIClient(cfg)

	// Make sure we have username
	if requestData.Username == "" {
		_AddBadRequestError(ww, "SendVerifyEmailEmail: No username sent")
		return
	}
	// Make sure we have link
	if requestData.Link == "" {
		_AddBadRequestError(ww, "SendVerifyEmailEmail: No link sent")
		return
	}
	// Make sure we have email
	if requestData.Email == "" {
		_AddBadRequestError(ww, "SendVerifyEmailEmail: No email sent in request")
		return
	}

	// Send email Structure
	var params interface{}
	params = map[string]interface{}{
        "username": requestData.Username,
		"link": requestData.Link,
    }

	// Make struct
	body := sendinblue.SendEmail{
		EmailTo:       []string{},
		EmailBcc:      []string{},
		EmailCc:       []string{},
		ReplyTo:       "",
		AttachmentUrl: "",
		Attachment:    []sendinblue.SendEmailAttachment{},
		Headers:       nil,
		Attributes:    nil,
		Tags:          []string{},
	}

	// Set Person who to send the email to
	body.EmailTo = []string{requestData.Email}
	// Set the map into attributes
	body.Attributes = &params

	// Send the email template
	_, _, err := sib.TransactionalEmailsApi.SendTemplate(ctx, body, 15)
	if err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("SendVerifyEmailEmail: Failed to send email: %v", err))
		return
	}
	// Return a Success response
	res := EmailSuccessResponse{
		Success: true,
	}
	if err = json.NewEncoder(ww).Encode(res); err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("SendVerifyEmailEmail: Problem encoding response as JSON: %v", err))
		return
	}
}
// Lost nft bid request
type LostNFTRequest struct {
	Username string `json:"username"`
	ArtName string `json:"artname"`
	Email string `json:"email"`
}
// Sent once user loses an nft to another bidder
func (fes *APIServer) SendLostNFTEmail(ww http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(io.LimitReader(req.Body, MaxRequestBodySizeBytes))
	// Check Request 
	requestData := LostNFTRequest{}
	if err := decoder.Decode(&requestData); err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("SendLostNFTEmail: Error parsing request body: %v", err))
		return
	}
	// Get context
	var ctx context.Context

	// Sendingblue configuration
	cfg := sendinblue.NewConfiguration()
	// Api key
	cfg.AddDefaultHeader("api-key", "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")
	// Partner key
	cfg.AddDefaultHeader("partner-key", "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")

	// Connect to SendinBlue Api client
	sib := sendinblue.NewAPIClient(cfg)

	// Make sure we have username
	if requestData.Username == "" {
		_AddBadRequestError(ww, "SendLostNFTEmail: No username sent")
		return
	}
	// Make sure we have art name
	if requestData.ArtName == "" {
		_AddBadRequestError(ww, "SendLostNFTEmail: No link sent")
		return
	}
	// Make sure we have email
	if requestData.Email == "" {
		_AddBadRequestError(ww, "SendLostNFTEmail: No email sent in request")
		return
	}

	// Send email Structure
	var params interface{}
	params = map[string]interface{}{
        "username": requestData.Username,
		"art_name": requestData.ArtName,
    }

	// Make struct
	body := sendinblue.SendEmail{
		EmailTo:       []string{},
		EmailBcc:      []string{},
		EmailCc:       []string{},
		ReplyTo:       "",
		AttachmentUrl: "",
		Attachment:    []sendinblue.SendEmailAttachment{},
		Headers:       nil,
		Attributes:    nil,
		Tags:          []string{},
	}

	// Set Person who to send the email to
	body.EmailTo = []string{requestData.Email}
	// Set the map into attributes
	body.Attributes = &params

	// Send the email template
	_, _, err := sib.TransactionalEmailsApi.SendTemplate(ctx, body, 15)
	if err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("SendLostNFTEmail: Failed to send email: %v", err))
		return
	}
	// Return a Success response
	res := EmailSuccessResponse{
		Success: true,
	}
	if err = json.NewEncoder(ww).Encode(res); err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("SendLostNFTEmail: Problem encoding response as JSON: %v", err))
		return
	}
}
// New bid made on nft
type NewBidRequest struct {
	CreatorUsername string `json:"creator_username"`
	BidderUsername string `json:"bidder_username"`
	BidAmount uint64 `json:"bid_amount"`
	LinkToNFT string `json:"link_to_nft"`
	Email string `json:"email"`
}
// Sent once someone makes a bid on users nft
func (fes *APIServer) SendNewBidEmail(ww http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(io.LimitReader(req.Body, MaxRequestBodySizeBytes))
	// Check Request 
	requestData := NewBidRequest{}
	if err := decoder.Decode(&requestData); err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("SendNewBidEmail: Error parsing request body: %v", err))
		return
	}
	// Get context
	var ctx context.Context

	// Sendingblue configuration
	cfg := sendinblue.NewConfiguration()
	// Api key
	cfg.AddDefaultHeader("api-key", "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")
	// Partner key
	cfg.AddDefaultHeader("partner-key", "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")

	// Connect to SendinBlue Api client
	sib := sendinblue.NewAPIClient(cfg)

	// Make sure we have CreatorUsername
	if requestData.CreatorUsername == "" {
		_AddBadRequestError(ww, "SendNewBidEmail: No CreatorUsername sent")
		return
	}
	// Make sure we have BidderUsername
	if requestData.BidderUsername == "" {
		_AddBadRequestError(ww, "SendNewBidEmail: No BidderUsername sent")
		return
	}
	// Make sure we have bid amount
	if requestData.BidAmount == 0 {
		_AddBadRequestError(ww, "SendNewBidEmail: No BidAmount sent")
		return
	}
	// Make sure we have a link to the nft
	if requestData.LinkToNFT == "" {
		_AddBadRequestError(ww, "SendNewBidEmail: No LinkToNFT sent:")
		return
	}
	// Make sure we have email
	if requestData.Email == "" {
		_AddBadRequestError(ww, "SendNewBidEmail: No email sent in request:")
		return
	}

	// Send email Structure
	var params interface{}
	params = map[string]interface{}{
        "creator_username": requestData.CreatorUsername,
		"bidder_username": requestData.BidderUsername,
		"bid_amount": requestData.BidAmount,
		"link_to_nft": requestData.LinkToNFT,
    }

	// Make struct
	body := sendinblue.SendEmail{
		EmailTo:       []string{},
		EmailBcc:      []string{},
		EmailCc:       []string{},
		ReplyTo:       "",
		AttachmentUrl: "",
		Attachment:    []sendinblue.SendEmailAttachment{},
		Headers:       nil,
		Attributes:    nil,
		Tags:          []string{},
	}

	// Set Person who to send the email to
	body.EmailTo = []string{requestData.Email}
	// Set the map into attributes
	body.Attributes = &params

	// Send the email template
	_, _, err := sib.TransactionalEmailsApi.SendTemplate(ctx, body, 15)
	if err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("SendNewBidEmail: Failed to send email: %v", err))
		return
	}
	// Return a Success response
	res := EmailSuccessResponse{
		Success: true,
	}
	if err = json.NewEncoder(ww).Encode(res); err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("SendNewBidEmail: Problem encoding response as JSON: %v", err))
		return
	}
}
// User has been inactive for a while
type InactiveUserRequest struct {
	Username string `json:"username"`
	LinkToProfile string `json:"link_to_profile"`
	Email string `json:"email"`
}
// Sent if a user has been inactive for long
func (fes *APIServer) SendInactiveUserEmail(ww http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(io.LimitReader(req.Body, MaxRequestBodySizeBytes))
	// Check Request 
	requestData := InactiveUserRequest{}
	if err := decoder.Decode(&requestData); err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("SendInactiveUserEmail: Error parsing request body: %v", err))
		return
	}
	// Get context
	var ctx context.Context

	// Sendingblue configuration
	cfg := sendinblue.NewConfiguration()
	// Api key
	cfg.AddDefaultHeader("api-key", "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")
	// Partner key
	cfg.AddDefaultHeader("partner-key", "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")

	// Connect to SendinBlue Api client
	sib := sendinblue.NewAPIClient(cfg)

	// Make sure we have Username
	if requestData.Username == "" {
		_AddBadRequestError(ww, "SendInactiveUserEmail: No Username sent")
		return
	}
	// Make sure we have link to profile
	if requestData.LinkToProfile == "" {
		_AddBadRequestError(ww, "SendInactiveUserEmail: No LinkToProfile sent")
		return
	}
	// Make sure we have email
	if requestData.Email == "" {
		_AddBadRequestError(ww, "SendInactiveUserEmail: No Email sent in request:")
		return
	}

	// Send email Structure
	var params interface{}
	params = map[string]interface{}{
        "username": requestData.Username,
		"link_to_profile": requestData.LinkToProfile,
    }

	// Make struct
	body := sendinblue.SendEmail{
		EmailTo:       []string{},
		EmailBcc:      []string{},
		EmailCc:       []string{},
		ReplyTo:       "",
		AttachmentUrl: "",
		Attachment:    []sendinblue.SendEmailAttachment{},
		Headers:       nil,
		Attributes:    nil,
		Tags:          []string{},
	}

	// Set Person who to send the email to
	body.EmailTo = []string{requestData.Email}
	// Set the map into attributes
	body.Attributes = &params

	// Send the email template
	_, _, err := sib.TransactionalEmailsApi.SendTemplate(ctx, body, 13)
	if err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("SendInactiveUserEmail: Failed to send email: %v", err))
		return
	}
	// Return a Success response
	res := EmailSuccessResponse{
		Success: true,
	}
	if err = json.NewEncoder(ww).Encode(res); err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("SendInactiveUserEmail: Problem encoding response as JSON: %v", err))
		return
	}
}
// Welcome email 
type WelcomeRequest struct {
	Username string `json:"username"`
	LinkToProfile string `json:"link_to_profile"`
	Email string `json:"email"`
}
// Sent when a user has onboarded
func (fes *APIServer) SendWelcomeEmail(ww http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(io.LimitReader(req.Body, MaxRequestBodySizeBytes))
	// Check Request 
	requestData := WelcomeRequest{}
	if err := decoder.Decode(&requestData); err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("SendWelcomeEmail: Error parsing request body: %v", err))
		return
	}
	// Get context
	var ctx context.Context

	// Sendingblue configuration
	cfg := sendinblue.NewConfiguration()
	// Api key
	cfg.AddDefaultHeader("api-key", "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")
	// Partner key
	cfg.AddDefaultHeader("partner-key", "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")

	// Connect to SendinBlue Api client
	sib := sendinblue.NewAPIClient(cfg)

	// Make sure we have Username
	if requestData.Username == "" {
		_AddBadRequestError(ww, "SendWelcomeEmail: No Username sent")
		return
	}
	// Make sure we have link to profile
	if requestData.LinkToProfile == "" {
		_AddBadRequestError(ww, "SendWelcomeEmail: No LinkToProfile sent")
		return
	}
	// Make sure we have email
	if requestData.Email == "" {
		_AddBadRequestError(ww, "SendWelcomeEmail: No Email sent in request")
		return
	}

	// Send email Structure
	var params interface{}
	params = map[string]interface{}{
        "username": requestData.Username,
		"link_to_profile": requestData.LinkToProfile,
    }

	// Make struct
	body := sendinblue.SendEmail{
		EmailTo:       []string{},
		EmailBcc:      []string{},
		EmailCc:       []string{},
		ReplyTo:       "",
		AttachmentUrl: "",
		Attachment:    []sendinblue.SendEmailAttachment{},
		Headers:       nil,
		Attributes:    nil,
		Tags:          []string{},
	}

	// Set Person who to send the email to
	body.EmailTo = []string{requestData.Email}
	// Set the map into attributes
	body.Attributes = &params

	// Send the email template
	_, _, err := sib.TransactionalEmailsApi.SendTemplate(ctx, body, 15)
	if err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("SendWelcomeEmail: Failed to send email: %v", err))
		return
	}
	// Return a Success response
	res := EmailSuccessResponse{
		Success: true,
	}
	if err = json.NewEncoder(ww).Encode(res); err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("SendWelcomeEmail: Problem encoding response as JSON: %v", err))
		return
	}
}
// Someone outbid the user
type BidAgainRequest struct {
	OutBiddedUsername string `json:"out_bidded_username"`
	OutBidderUsername string `json:"out_bidder_username"`
	NewBidAmount int64 `json:"new_bid_amount"`
	LinkToNFT string `json:"link_to_nft"`
	Email string `json:"email"`
}
// Sent when a user has been outbidded
func (fes *APIServer) SendBidAgainEmail(ww http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(io.LimitReader(req.Body, MaxRequestBodySizeBytes))
	// Check Request 
	requestData := BidAgainRequest{}
	if err := decoder.Decode(&requestData); err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("SendBidAgainEmail: Error parsing request body: %v", err))
		return
	}
	// Get context
	var ctx context.Context

	// Sendingblue configuration
	cfg := sendinblue.NewConfiguration()
	// Api key
	cfg.AddDefaultHeader("api-key", "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")
	// Partner key
	cfg.AddDefaultHeader("partner-key", "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")

	// Connect to SendinBlue Api client
	sib := sendinblue.NewAPIClient(cfg)

	// Make sure we have OutBiddedUsername
	if requestData.OutBiddedUsername == "" {
		_AddBadRequestError(ww, "SendBidAgainEmail: No OutBiddedUsername sent")
		return
	}
	// Make sure we have link to OutBidderUsername 
	if requestData.OutBidderUsername  == "" {
		_AddBadRequestError(ww, "SendBidAgainEmail: No OutBidderUsername sent")
		return
	}
	// Make sure we have link to NewBidAmount
	if requestData.NewBidAmount  == 0 {
		_AddBadRequestError(ww, "SendBidAgainEmail: No NewBidAmount sent")
		return
	}
	// Make sure we have link to LinkToNFT
	if requestData.LinkToNFT  == "" {
		_AddBadRequestError(ww, "SendBidAgainEmail: No LinkToNFT sent")
		return
	}
	// Make sure we have email
	if requestData.Email == "" {
		_AddBadRequestError(ww, "SendBidAgainEmail: No Email sent in request")
		return
	}

	// Send email Structure
	var params interface{}
	params = map[string]interface{}{
        "out_bidded_username": requestData.OutBiddedUsername,
		"out_bidder_username": requestData.OutBidderUsername,
		"new_bid_amount": requestData.NewBidAmount,
		"link_to_nft": requestData.LinkToNFT,

    }

	// Make struct
	body := sendinblue.SendEmail{
		EmailTo:       []string{},
		EmailBcc:      []string{},
		EmailCc:       []string{},
		ReplyTo:       "",
		AttachmentUrl: "",
		Attachment:    []sendinblue.SendEmailAttachment{},
		Headers:       nil,
		Attributes:    nil,
		Tags:          []string{},
	}

	// Set Person who to send the email to
	body.EmailTo = []string{requestData.Email}
	// Set the map into attributes
	body.Attributes = &params

	// Send the email template
	_, _, err := sib.TransactionalEmailsApi.SendTemplate(ctx, body, 15)
	if err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("SendBidAgainEmail: Failed to send email: %v", err))
		return
	}
	// Return a Success response
	res := EmailSuccessResponse{
		Success: true,
	}
	if err = json.NewEncoder(ww).Encode(res); err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("SendBidAgainEmail: Problem encoding response as JSON: %v", err))
		return
	}
}
// Won NFT
type WonNFTRequest struct {
	WinnerUsername string `json:"winner_username"`
	ArtName string `json:"art_name"`
	WinningBidAmount uint64 `json:"winning_bid_amount"`
	LinkToNFT string `json:"link_to_nft"`
	Email string `json:"email"`
}
// Sent when a user has the winning bid
func (fes *APIServer) SendWonNFTEmail(ww http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(io.LimitReader(req.Body, MaxRequestBodySizeBytes))
	// Check Request 
	requestData := WonNFTRequest{}
	if err := decoder.Decode(&requestData); err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("SendWonNFTEmail: Error parsing request body: %v", err))
		return
	}
	// Get context
	var ctx context.Context

	// Sendingblue configuration
	cfg := sendinblue.NewConfiguration()
	// Api key
	cfg.AddDefaultHeader("api-key", "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")
	// Partner key
	cfg.AddDefaultHeader("partner-key", "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")

	// Connect to SendinBlue Api client
	sib := sendinblue.NewAPIClient(cfg)

	// Make sure we have WinnerUsername
	if requestData.WinnerUsername == "" {
		_AddBadRequestError(ww, "SendWonNFTEmail: No WinnerUsername sent")
		return
	}
	// Make sure we have link to ArtName
	if requestData.ArtName == "" {
		_AddBadRequestError(ww, "SendWonNFTEmail: No ArtName sent")
		return
	}
	// Make sure we have link to WinningBidAmount
	if requestData.WinningBidAmount == 0 {
		_AddBadRequestError(ww, "SendWonNFTEmail: No WinningBidAmount sent")
		return
	}
	// Make sure we have link to LinkToNFT
	if requestData.LinkToNFT == "" {
		_AddBadRequestError(ww, "SendWonNFTEmail: No LinkToNFT sent")
		return
	}
	// Make sure we have email
	if requestData.Email == "" {
		_AddBadRequestError(ww, "SendWonNFTEmail: No Email sent in request")
		return
	}

	// Send email Structure
	var params interface{}
	params = map[string]interface{}{
        "winner_username": requestData.WinnerUsername,
		"art_name": requestData.ArtName,
		"winning_bid_amount": requestData.WinningBidAmount,
		"link_to_nft": requestData.LinkToNFT,

    }

	// Make struct
	body := sendinblue.SendEmail{
		EmailTo:       []string{},
		EmailBcc:      []string{},
		EmailCc:       []string{},
		ReplyTo:       "",
		AttachmentUrl: "",
		Attachment:    []sendinblue.SendEmailAttachment{},
		Headers:       nil,
		Attributes:    nil,
		Tags:          []string{},
	}

	// Set Person who to send the email to
	body.EmailTo = []string{requestData.Email}
	// Set the map into attributes
	body.Attributes = &params

	// Send the email template
	_, _, err := sib.TransactionalEmailsApi.SendTemplate(ctx, body, 10)
	if err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("SendBidAgainEmail: Failed to send email: %v", err))
		return
	}
	// Return a Success response
	res := EmailSuccessResponse{
		Success: true,
	}
	if err = json.NewEncoder(ww).Encode(res); err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("SendBidAgainEmail: Problem encoding response as JSON: %v", err))
		return
	}
}
// NFT bid place
type BidPlaceRequest struct {
	Username string `json:"username"`
	BidAmount uint64 `json:"bid_amount"`
	LinkToNFT string `json:"link_to_nft"`
	Email string `json:"email"`
}
// Sent when a user has the winning bid
func (fes *APIServer) SendBidPlacedEmail(ww http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(io.LimitReader(req.Body, MaxRequestBodySizeBytes))
	// Check Request 
	requestData := BidPlaceRequest{}
	if err := decoder.Decode(&requestData); err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("SendBidPlaceEmail: Error parsing request body: %v", err))
		return
	}
	// Get context
	var ctx context.Context

	// Sendingblue configuration
	cfg := sendinblue.NewConfiguration()
	// Api key
	cfg.AddDefaultHeader("api-key", "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")
	// Partner key
	cfg.AddDefaultHeader("partner-key", "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")

	// Connect to SendinBlue Api client
	sib := sendinblue.NewAPIClient(cfg)

	// Make sure we have Username
	if requestData.Username == "" {
		_AddBadRequestError(ww, "SendBidPlaceEmail: No Username sent")
		return
	}
	// Make sure we have link to BidAmount
	if requestData.BidAmount == 0 {
		_AddBadRequestError(ww, "SendBidPlaceEmail: No BidAmount sent")
		return
	}
	// Make sure we have link to LinkToNFT
	if requestData.LinkToNFT == "" {
		_AddBadRequestError(ww, "SendBidPlaceEmail: No LinkToNFT sent")
		return
	}
	// Make sure we have email
	if requestData.Email == "" {
		_AddBadRequestError(ww, "SendBidPlaceEmail: No Email sent in request")
		return
	}

	// Send email Structure
	var params interface{}
	params = map[string]interface{}{
        "username": requestData.Username,
		"bid_amount": requestData.BidAmount,
		"link_to_nft": requestData.LinkToNFT,
    }

	// Make struct
	body := sendinblue.SendEmail{
		EmailTo:       []string{},
		EmailBcc:      []string{},
		EmailCc:       []string{},
		ReplyTo:       "",
		AttachmentUrl: "",
		Attachment:    []sendinblue.SendEmailAttachment{},
		Headers:       nil,
		Attributes:    nil,
		Tags:          []string{},
	}

	// Set Person who to send the email to
	body.EmailTo = []string{requestData.Email}
	// Set the map into attributes
	body.Attributes = &params

	// Send the email template
	_, _, err := sib.TransactionalEmailsApi.SendTemplate(ctx, body, 13)
	if err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("SendBidPlaceEmail: Failed to send email: %v", err))
		return
	}
	if err = json.NewEncoder(ww).Encode(body); err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("SendBidPlaceEmail: Problem encoding response as JSON: %v", err))
		return
	}
}
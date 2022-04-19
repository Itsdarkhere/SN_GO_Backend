package routes

import (
	"context"
	"encoding/json"
	"net/http"
	"io"
	"fmt"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// Used for all routes as the success response
type EmailSuccessResponse struct {
	Success bool `json:"success"`
}

type SendWelcomeEmailRequest struct {
	Username string `json:"Username"`
	Email string `json:"Email"`
}

func (fes *APIServer) SendWelcomeEmail(ww http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(io.LimitReader(req.Body, MaxRequestBodySizeBytes))
	// Check Request 
	requestData := SendWelcomeEmailRequest{}
	if err := decoder.Decode(&requestData); err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("SendWelcomeEmail: Error parsing request body: %v", err))
		return
	}

	if requestData.Username == "" {
		_AddBadRequestError(ww, fmt.Sprintf("SendWelcomeEmail: Error no Username sent in request"))
		return
	}

	if requestData.Email == "" {
		_AddBadRequestError(ww, fmt.Sprintf("SendWelcomeEmail: Error no Email sent in request"))
		return
	}
	// Username, Email
	m := mail.NewV3Mail()

	address := "support@supernovas.app"
	name := "Supernovas"
	e := mail.NewEmail(name, address)
	m.SetFrom(e)

	m.SetTemplateID("d-6c7890481de049f49c73bd838fe8941f")

	p := mail.NewPersonalization()
	tos := []*mail.Email{
		mail.NewEmail(requestData.Username, requestData.Email),
	}
	p.AddTos(tos...)

  	m.AddPersonalizations(p)
	
	request := sendgrid.GetRequest("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", "/v3/mail/send", "https://api.sendgrid.com")
	request.Method = "POST"
	var Body = mail.GetRequestBody(m)
	request.Body = Body
	response, err := sendgrid.API(request)
	if err = json.NewEncoder(ww).Encode("Success"); err != nil {
		_AddInternalServerError(ww, fmt.Sprintf("SendWelcomeEmail: Error sending email: %v", err))
		return
	}
}

type AddToInvestorEmailListRequest struct {
	Email string 
}
// Adds person to investors info emails list
func (fes *APIServer) AddToInvestorEmailList(ww http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(io.LimitReader(req.Body, MaxRequestBodySizeBytes))
	// Check Request 
	requestData := AddToInvestorEmailListRequest{}
	if err := decoder.Decode(&requestData); err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("AddToInvestorEmailList: Error parsing request body: %v", err))
		return
	}

	if requestData.Email == "" {
		_AddBadRequestError(ww, fmt.Sprintf("AddToInvestorEmailList: Error no Email sent in request"))
		return
	}

	host := "https://api.sendgrid.com"
    request := sendgrid.GetRequest("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", "/v3/marketing/contacts", host)
    request.Method = "PUT"
    request.Body = []byte(fmt.Sprintf(`{
		"contacts": [
			{
			"email": "%v"
			}
		],
		"list_ids": [
			"f7664252-5394-48b6-86cd-af1a683d928d"
		]
		}`, requestData.Email))

    response, err := sendgrid.API(request)
    if err = json.NewEncoder(ww).Encode("Success"); err != nil {
		_AddInternalServerError(ww, fmt.Sprintf("AddToInvestorEmailList: Problem serializing object to JSON: %v", err))
		return
	}
}
type ReportPostEmailRequest struct {
	ReportedContent string `json:"ReportedContent"`
}

func (fes *APIServer) ReportPostEmail(ww http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(io.LimitReader(req.Body, MaxRequestBodySizeBytes))
	// Check Request 
	requestData := ReportPostEmailRequest{}
	if err := decoder.Decode(&requestData); err != nil {
		_AddBadRequestError(ww, fmt.Sprintf("ReportPostEmail: Error parsing request body: %v", err))
		return
	}

	if requestData.ReportedContent == "" {
		_AddBadRequestError(ww, fmt.Sprintf("ReportPostEmail: Error no ReportedContent sent in request"))
		return
	}

	from := mail.NewEmail("Supernovas", "support@supernovas.app")
	subject := "Reported content"
	to := mail.NewEmail("Support", "support@supernovas.app")
	plainTextContent := "Content reported:"
	link := requestData.ReportedContent
	htmlContent := link
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	client := sendgrid.NewSendClient("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")
	response, err := client.Send(message)

	if err = json.NewEncoder(ww).Encode("Success"); err != nil {
		_AddInternalServerError(ww, fmt.Sprintf("ReportPostEmail: Problem serializing object to JSON: %v", err))
		return
	}
}
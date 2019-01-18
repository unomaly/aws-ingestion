package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// Event expected by the HTTP ingestion API
type UnomalyEvent struct {
	Message   string                 `json:"message"`
	Source    string                 `json:"source"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// Lambda response struct
type Response struct {
	Ok      bool
	Message string
}

var unomalyHost, lambdaName string
var acceptSelfSignedCerts, keepTimestamp bool
var batchSize int64
var unomalyBatch []*UnomalyEvent

// Post a batch of events to unomaly
func postBatch() error {
	if len(unomalyBatch) == 0 {
		return nil
	}

	data, err := json.Marshal(unomalyBatch)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON payload: %s", err)

	}

	resp, err := http.Post(unomalyHost+"/v1/batch", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	err = resp.Body.Close()

	if err != nil {
		return fmt.Errorf("failed to marshal JSON payload: %s", err)
	}
	unomalyBatch = make([]*UnomalyEvent, 0)
	return nil
}

// Lambda handler
func handler(request events.CloudwatchLogsEvent) (Response, error) {
	unomalyBatch = make([]*UnomalyEvent, 0)
	d, err := request.AWSLogs.Parse()

	if err != nil {
		log.Printf("Failed to parse event: %s", err)
		return Response{
			Ok:      false,
			Message: fmt.Sprintf("Failed to parse event: %s", err),
		}, err
	}

	for _, event := range d.LogEvents {

		unomalyEv := &UnomalyEvent{
			Message:  event.Message,
			Source:   d.LogGroup,
			Metadata: make(map[string]interface{}),
		}

		if keepTimestamp {
			unomalyEv.Timestamp = time.Unix(0, event.Timestamp*1000000)
		}

		unomalyEv.Metadata["log_stream"] = d.LogStream
		unomalyEv.Metadata["log_group"] = d.LogGroup
		unomalyEv.Metadata["message_type"] = d.MessageType
		unomalyEv.Metadata["owner"] = d.Owner
		unomalyEv.Metadata["subscription_filters"] = d.SubscriptionFilters
		unomalyEv.Metadata["unomaly_input"] = "cloudwatch_lamda"
		unomalyEv.Metadata["lambda_name"] = lambdaName

		unomalyBatch = append(unomalyBatch, unomalyEv)

		// Publish batch if necessary
		if len(unomalyBatch) == int(batchSize) {
			err = postBatch()
			if err != nil {
				log.Printf("Failed to post batch: %s", err)
				return Response{
					Ok:      false,
					Message: fmt.Sprintf("Failed to post batch: %s", err),
				}, err
			}

		}
	}

	err = postBatch()
	if err != nil {
		return Response{
			Ok:      false,
			Message: fmt.Sprintf("Failed to post batch: %s", err),
		}, err
	}

	return Response{
		Ok:      true,
		Message: fmt.Sprintf("%d events sent", len(unomalyBatch)),
	}, nil
}

func main() {
	lambdaName = os.Getenv("AWS_LAMBDA_FUNCTION_NAME")

	acceptSelfSigned := os.Getenv("ACCEPT_SELF_SIGNED_CERTS")

	if acceptSelfSigned == "true" {
		acceptSelfSignedCerts = true
	}

	keepTimestampString := os.Getenv("KEEP_TIMESTAMP")

	keepTimestamp = true
	if keepTimestampString == "false" {
		keepTimestamp = false
	}

	batchSizeString := os.Getenv("BATCH_SIZE")

	if batchSizeString != "" {
		var err error
		batchSize, err = strconv.ParseInt(batchSizeString, 10, 64)
		if err != nil {
			log.Fatalf("Can not parse batch size: %s", err)
		}
	} else {
		batchSize = 100
	}

	unomalyHost = os.Getenv("UNOMALY_HOST")
	if unomalyHost == "" {
		log.Fatalf("Missing UNOMALY_HOST env var")
	}

	if !strings.HasPrefix(unomalyHost, "http") {
		unomalyHost = "https://" + unomalyHost
	}
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: acceptSelfSignedCerts}

	lambda.Start(handler)
}

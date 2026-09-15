package trip

import (
	"bombe_main_server/internal/awssqs"
	redis_handler "bombe_main_server/internal/redis"
	"bombe_main_server/internal/websocket"
	socket_model "bombe_main_server/internal/websocket/model"
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
)

type TripHandler struct {
	redis *redis_handler.RedisHandler
	sqs   *awssqs.SQSQueue
	ws    *websocket.WSHandlerS
}

func NewTripHandler(
	redis *redis_handler.RedisHandler,
	sqs *awssqs.SQSQueue,
	ws *websocket.WSHandlerS,
) *TripHandler {
	return &TripHandler{
		redis: redis,
		sqs:   sqs,
		ws:    ws,
	}
}

func (t *TripHandler) ReadMessagesStoredInChannel() {

	type WSMessage struct {
		Type string `json:"type"`
	}

	for msg := range t.ws.MsgCh {

		var wms WSMessage

		if err := json.Unmarshal(msg, &wms); err != nil {
			fmt.Println("error parsing json while reading from channel:", err)
			continue
		}

		switch wms.Type {

		case "LOCATION_UPDATE_FROM_DRIVER":

			var data socket_model.DataToSendOnLive

			if err := json.Unmarshal(msg, &data); err != nil {
				fmt.Println("Invalid JSON:", err)
				continue
			}

			err := t.redis.SetDriverOnline(
				context.Background(),
				data.DriverID,
				data.Latitude,
				data.Longitude,
			)

			if err != nil {
				fmt.Println("Unable to set driver online:", err)
			}

			if data.RiderID != "" && data.RideID != "" {

				err = t.redis.RemoveDriver(
					context.Background(),
					data.DriverID,
				)

				if err != nil {
					fmt.Println("removing driver from redis error:", err)
				}

				err = t.ws.SendToTargetUser(
					data.RiderID,
					socket_model.DataToSendTORiderFromDriver{
						DriverID:  data.DriverID,
						Type:      "ASSIGNED_DRIVER_LOCATION_UPDATE",
						RiderID:   data.RiderID,
						RideID:    data.RideID,
						Latitude:  data.Latitude,
						Longitude: data.Longitude,
						Status:    data.Status,
					},
				)

				if err != nil {
					fmt.Println("Unable to send data to rider:", err)
				}
			}

		default:
			fmt.Println("message with no assigned type")
			fmt.Println(string(msg))
		}
	}
}

// sending trip request to drivers
func (t *TripHandler) StartListeningToSQS() {

	fmt.Println("start listening sqs")

	for {

		messages, err := t.sqs.ReceiveMessages(context.Background())

		if err != nil {
			fmt.Printf("error start sqs: %s\n", err.Error())
			continue
		}

		for _, message := range messages {

			if message.Body == nil {
				continue
			}

			fmt.Printf(
				"Message ID: %s\n",
				aws.ToString(message.MessageId),
			)

			var data SQSDataFor

			err := json.Unmarshal(
				[]byte(*message.Body),
				&data,
			)

			fmt.Printf("Message ID: %s\n", aws.ToString(message.MessageId))

			fmt.Println("========== RAW SQS BODY ==========")
			fmt.Println(*message.Body)
			fmt.Println("==================================")

			if err != nil {
				fmt.Println(
					"error while parsing json for FOR:",
					err,
				)
				continue
			}

			switch data.For {

			case "TRIP_CREATED":

				var tripData SQSDataToMainServer

				err := json.Unmarshal(
					[]byte(*message.Body),
					&tripData,
				)

				if err != nil {
					fmt.Printf(
						"invalid SQS message: %s\n",
						err.Error(),
					)
					continue
				}

				if tripData.Aud != "MAIN_SERVER" {
					continue
				}

				driverIDs, err := t.redis.FindDriversNearby(
					context.Background(),
					tripData.Pickup.Latitude,
					tripData.Pickup.Longitude,
					5,
				)

				if err != nil {
					fmt.Println(
						"error finding drivers:",
						err,
					)
					continue
				}

				for _, driverID := range driverIDs {

					fmt.Println(
						"total drivers found:",
						len(driverIDs),
						"driver:",
						driverID,
					)

					err := t.ws.SendToTargetUser(
						driverID,
						TripReqToDriver{
							Type:        "RIDE_REQUEST_TO_DRIVER",
							DriverFare:  tripData.DriverFare,
							RideDetails: tripData.RideDetails,
							Pickup:      tripData.Pickup,
							Dropoff:     tripData.Dropoff,
							RiderID:     tripData.RiderID,
							TripID:      tripData.TripID,
						},
					)

					if err != nil {
						fmt.Println(
							"error sending trip request to driver:",
							err,
						)
					}
				}

				if err := t.sqs.DeleteMessage(
					context.Background(),
					*message.ReceiptHandle,
				); err != nil {
					fmt.Println(
						"error deleting message, will try again:",
						err,
					)
				}

			case "CANCEL_TRIP":

				var tripData SQSTripCancelRequestToMain

				err := json.Unmarshal(
					[]byte(*message.Body),
					&tripData,
				)

				if err != nil {
					fmt.Printf(
						"invalid SQS message trip cancel: %s\n",
						err.Error(),
					)
					continue
				}

				data := CancelTripDataToUser{
					Type:     "CANCEL_TRIP",
					TripId:   tripData.TripId,
					RiderID:  tripData.RiderId,
					DriverId: tripData.DriverId,
					Message:  "coming soon",
				}

				if tripData.IsCancelledByDriver {

					// Driver cancelled -> notify rider
					data.DriverOrRider = "rider"

					err = t.ws.SendToTargetUser(
						tripData.RiderId,
						data,
					)

				} else {

					// Rider cancelled -> notify driver
					data.DriverOrRider = "driver"

					err = t.ws.SendToTargetUser(
						tripData.DriverId,
						data,
					)
				}

				if err != nil {
					fmt.Println(
						"error sending cancel trip message to user:",
						err,
					)
				}

				if err := t.sqs.DeleteMessage(
					context.Background(),
					*message.ReceiptHandle,
				); err != nil {
					fmt.Println(
						"error deleting cancel trip message:",
						err,
					)
				}

			case "TRIP_COMPLETE":

				var tripData SQSTripCompletedRequestToMain

				err := json.Unmarshal(
					[]byte(*message.Body),
					&tripData,
				)

				if err != nil {
					fmt.Printf(
						"invalid SQS message trip complete: %s\n",
						err.Error(),
					)
					continue
				}

				data := CompleteTripDataToUser{
					Type:     "COMPLETE_TRIP",
					TripId:   tripData.TripId,
					DriverId: tripData.DriverId,
					RiderId:  tripData.RiderId,
				}

				if err := t.ws.SendToTargetUser(
					tripData.RiderId,
					data,
				); err != nil {
					fmt.Println(
						"error sending trip complete message:",
						err,
					)
				}

				if err := t.sqs.DeleteMessage(
					context.Background(),
					*message.ReceiptHandle,
				); err != nil {
					fmt.Println(
						"error deleting trip complete message:",
						err,
					)
				}

			default:

				fmt.Println(
					"unknown SQS message type:",
					data.For,
				)

				continue
			}
		}
	}
}

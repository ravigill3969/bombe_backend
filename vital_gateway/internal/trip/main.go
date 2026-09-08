package trip

import (
	"bombe_main_server/internal/awssqs"
	redis_handler "bombe_main_server/internal/redis"
	"bombe_main_server/internal/websocket"
	socket_model "bombe_main_server/internal/websocket/model"
	"context"
	"encoding/json"
	"fmt"
)

type TripHandler struct {
	redis *redis_handler.RedisHandler
	sqs   *awssqs.SQSQueue
	ws    *websocket.WSHandlerS
}

func NewTripHandler(redis *redis_handler.RedisHandler, sqs *awssqs.SQSQueue, ws *websocket.WSHandlerS) *TripHandler {
	return &TripHandler{
		redis: redis,
		sqs:   sqs,
		ws:    ws,
	}
}

func (t *TripHandler) ReadMessagesStoredInChannel() {

	for msg := range t.ws.MsgCh {

		var data socket_model.DataToSendOnLive

		if err := json.Unmarshal(msg, &data); err != nil {
			fmt.Println("Invalid JSON:", err)
			continue
		}

		err := t.redis.SetDriverOnline(context.Background(), data.DriverID, data.Latitude, data.Longitude)

		if err != nil {
			fmt.Println("Unable to set driver online", err)
		}

		if data.RiderID != "" && data.RideID != "" {

			err = t.redis.RemoveDriver(context.Background(), data.DriverID)
			if err != nil {
				fmt.Println("removing rider from redis error", err)
			}

			err = t.ws.SendToTargetUser(data.RiderID, socket_model.DataToSendTORiderFromDriver{
				DriverID:  data.DriverID,
				Type:      "ASSIGNED_DRIVER_LOCATION_UPDATE",
				RiderID:   data.RiderID,
				RideID:    data.RideID,
				Latitude:  data.Latitude,
				Longitude: data.Longitude,
				Status: data.Status,
			})

			if err != nil {
				fmt.Println("Unable to send data to rider", err)
			}
		}
		// fmt.Println(err)
	}

}

// sending trip req to driver
func (t *TripHandler) StartListeningToSQS() {
	fmt.Println("start listeing sqs")
	for {

		messages, err := t.sqs.ReceiveMessages(context.Background())

		if err != nil {
			fmt.Printf("error start sqs %s", err.Error())
			continue
		}

		for _, message := range messages {
			if message.Body == nil {
				continue
			}

			var data SQSDataToMainServer

			err := json.Unmarshal(
				[]byte(*message.Body),
				&data,
			)

			if err != nil {
				fmt.Printf("invalid SQS message: %s", err.Error())
				continue
			}

			if data.Aud != "MAIN_SERVER" {
				continue
			}

			driver_ids, err := t.redis.FindDriversNearby(context.Background(), data.Pickup.Latitude, data.Pickup.Longitude, 5)

			if err != nil {
				fmt.Println("error finding drivers ", err)
			}


			for _, e := range driver_ids {
				fmt.Println(e)
				t.ws.SendToTargetUser(e, TripReqToDriver{
					Type:        "RIDE_REQUEST_TO_DRIVER",
					DriverFare:  data.DriverFare,
					RideDetails: data.RideDetails,
					Pickup:      data.Pickup,
					Dropoff:     data.Dropoff,
					RiderID:     data.RiderID,
					TripID:      data.TripID,
				})
			}

			err = t.sqs.DeleteMessage(context.Background(), *message.ReceiptHandle)

			if err != nil {
				fmt.Println("error deleting message will try again: ", err)
			}
		}

	}
}

//sending driver location to rider
// read from ws
// send to redis and rider

func ( t *TripHandler)SendDriverLocationInfoToRider(){
	
}
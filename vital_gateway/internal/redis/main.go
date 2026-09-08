package redis_handler

import (
	"context"

	"github.com/redis/go-redis/v9"
)

const driverLocationKey = "drivers:locations"

type RedisHandler struct {
	redisC *redis.Client
}

func NewRedisHandler(redisURL string) *RedisHandler {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		panic(err)
	}

	client := redis.NewClient(opt)

	return &RedisHandler{
		redisC: client,
	}
}

type DriverLocation struct {
	DriverID  string
	Latitude  float64
	Longitude float64
	IsOnline  bool
}

func (r *RedisHandler) SetDriverLocation(
	ctx context.Context,
	data DriverLocation,
) error {
	if !data.IsOnline {
		return r.RemoveDriver(ctx, data.DriverID)
	}

	return r.redisC.GeoAdd(ctx, driverLocationKey, &redis.GeoLocation{
		Name:      data.DriverID,
		Longitude: data.Longitude,
		Latitude:  data.Latitude,
	}).Err()
}

func (r *RedisHandler) RemoveDriver(
	ctx context.Context,
	driverID string,
) error {
	return r.redisC.ZRem(
		ctx,
		driverLocationKey,
		driverID,
	).Err()
}

func (r *RedisHandler) FindDriversNearby(
	ctx context.Context,
	latitude float64,
	longitude float64,
	radiusKm float64,
) ([]string, error) {

	result, err := r.redisC.GeoSearch(
		ctx,
		driverLocationKey,
		&redis.GeoSearchQuery{
			Longitude:  longitude,
			Latitude:   latitude,
			Radius:     radiusKm,
			RadiusUnit: "km",
			Sort:       "ASC",
		},
	).Result()

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *RedisHandler) SetDriverOnline(
	ctx context.Context,
	driverID string,
	latitude float64,
	longitude float64,
) error {

	pipe := r.redisC.TxPipeline()

	pipe.GeoAdd(ctx, "drivers:locations", &redis.GeoLocation{
		Name:      driverID,
		Longitude: longitude,
		Latitude:  latitude,
	})

	pipe.SAdd(ctx, "drivers:available", driverID)

	pipe.HSet(ctx, "driver:"+driverID,
		"status", "available",
		"ride_id", "",
	)

	_, err := pipe.Exec(ctx)
	return err
}

func (r *RedisHandler) AssignDriver(
	ctx context.Context,
	driverID string,
	rideID string,
) error {

	pipe := r.redisC.TxPipeline()

	pipe.SRem(ctx, "drivers:available", driverID)

	pipe.HSet(ctx, "driver:"+driverID,
		"status", "assigned",
		"ride_id", rideID,
	)

	_, err := pipe.Exec(ctx)
	return err
}

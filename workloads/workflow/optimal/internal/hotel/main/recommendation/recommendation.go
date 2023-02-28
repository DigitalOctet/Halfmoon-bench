package recommendation

import (
	"math"

	"github.com/eniac/Beldi/internal/hotel/main/data"
	"github.com/eniac/Beldi/pkg/cayonlib"
	"github.com/hailocab/go-geoindex"
	"github.com/mitchellh/mapstructure"
)

func LoadRecommendations(env *cayonlib.Env) []data.Recommend {
	var recommends []data.Recommend
	res := cayonlib.Scan(env, data.Trecommendation())
	if res == nil {
		return []data.Recommend{}
	}
	err := mapstructure.Decode(res, &recommends)
	cayonlib.CHECK(err)
	return recommends
}

func GetRecommendations(env *cayonlib.Env, req Request) Result {
	hotels := LoadRecommendations(env)
	res := Result{HotelIds: []string{}}
	switch req.Require {
	case "dis":
		p1 := &geoindex.GeoPoint{
			Pid:  "",
			Plat: req.Lat,
			Plon: req.Lon,
		}
		min := math.MaxFloat64
		for _, hotel := range hotels {
			tmp := float64(geoindex.Distance(p1, &geoindex.GeoPoint{
				Pid:  "",
				Plat: hotel.HLat,
				Plon: hotel.HLon,
			})) / 1000
			if tmp < min {
				min = tmp
			}
		}
		for _, hotel := range hotels {
			tmp := float64(geoindex.Distance(p1, &geoindex.GeoPoint{
				Pid:  "",
				Plat: hotel.HLat,
				Plon: hotel.HLon,
			})) / 1000
			if tmp == min {
				res.HotelIds = append(res.HotelIds, hotel.HId)
			}
		}
	case "rate":
		max := 0.0
		for _, hotel := range hotels {
			if hotel.HRate > max {
				max = hotel.HRate
			}
		}
		for _, hotel := range hotels {
			if hotel.HRate == max {
				res.HotelIds = append(res.HotelIds, hotel.HId)
			}
		}
	case "price":
		min := math.MaxFloat64
		for _, hotel := range hotels {
			if hotel.HPrice < min {
				min = hotel.HPrice
			}
		}
		for _, hotel := range hotels {
			if hotel.HPrice == min {
				res.HotelIds = append(res.HotelIds, hotel.HId)
			}
		}
	default:
		panic("no such requirement")
	}
	return res
}

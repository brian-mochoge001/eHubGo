package handlers

import (
	"context"
	"os"
	"strconv"

	"github.com/gojuno/go.osrm"
	"github.com/paulmach/go.geo"
	polyline "github.com/twpayne/go-polyline"
)

var osrmSvc *osrm.OSRM

func init() {
	osrmURL := os.Getenv("OSRM_URL")
	if osrmURL == "" {
		osrmURL = "https://router.project-osrm.org/"
	}
	osrmSvc = osrm.NewFromURL(osrmURL)
}

// EncodeSimplePath creates a polyline string from a set of lat/lng pairs.
func EncodeSimplePath(coords [][]float64) string {
	if len(coords) == 0 {
		return ""
	}
	return string(polyline.EncodeCoords(coords))
}

// EncodePointToPoint creates a polyline for a straight line between two points.
func EncodePointToPoint(lat1, lng1, lat2, lng2 float64) string {
	coords := [][]float64{
		{lat1, lng1},
		{lat2, lng2},
	}
	return EncodeSimplePath(coords)
}

// GetFastestRoutePolyline computes the road-network path between multiple points.
func GetFastestRoutePolyline(ctx context.Context, points [][]float64) string {
	if len(points) < 2 {
		return ""
	}

	pointSet := geo.PointSet{}
	for _, p := range points {
		pointSet = append(pointSet, geo.Point{p[1], p[0]})
	}

	req := osrm.RouteRequest{
		Profile:     "car",
		Coordinates: osrm.NewGeometryFromPointSet(pointSet),
		Overview:    osrm.OverviewFull,
		Geometries:  osrm.GeometriesPolyline6, // Using Polyline6 as fallback or if available
		Steps:       osrm.StepsTrue,
	}

	res, err := osrmSvc.Route(ctx, req)
	if err != nil || len(res.Routes) == 0 || len(res.Routes[0].Legs) == 0 {
		return EncodeSimplePath(points)
	}

	// Manual assembly since OverviewFull might not populate a single route geometry in some library versions
	mergedPointSet := geo.PointSet{}
	for _, leg := range res.Routes[0].Legs {
		for _, step := range leg.Steps {
			mergedPointSet = append(mergedPointSet, step.Geometry.Points()...)
		}
	}
	
	if len(mergedPointSet) > 0 {
		path := geo.NewPath()
		path.SetPoints(mergedPointSet)
		return path.Encode()
	}

	return EncodeSimplePath(points)
}

// GetPointToPointRoute computes the fastest road route between two locations.
func GetPointToPointRoute(ctx context.Context, lat1, lng1, lat2, lng2 float64) string {
	return GetFastestRoutePolyline(ctx, [][]float64{{lat1, lng1}, {lat2, lng2}})
}

// SnapToRoad finds the nearest navigable road point for a given GPS coordinate.
func SnapToRoad(ctx context.Context, lat, lng float64) (float64, float64) {
	req := osrm.NearestRequest{
		Profile:     "car",
		Coordinates: osrm.NewGeometryFromPointSet(geo.PointSet{{lng, lat}}),
		Number:      1,
	}

	res, err := osrmSvc.Nearest(ctx, req)
	if err != nil || len(res.Waypoints) == 0 {
		return lat, lng
	}

	// OSRM returns [lng, lat]
	return res.Waypoints[0].Location[1], res.Waypoints[0].Location[0]
}

// GetDurationsMatrix computes a duration matrix between sources and destinations.
// Useful for finding the single closest driver among many.
func GetDurationsMatrix(ctx context.Context, sources, destinations [][]float64) [][]float64 {
	allPoints := geo.PointSet{}
	sourceIndices := []int{}
	destIndices := []int{}

	for i, p := range sources {
		allPoints = append(allPoints, geo.Point{p[1], p[0]})
		sourceIndices = append(sourceIndices, i)
	}

	offset := len(sources)
	for i, p := range destinations {
		allPoints = append(allPoints, geo.Point{p[1], p[0]})
		destIndices = append(destIndices, offset+i)
	}

	req := osrm.TableRequest{
		Profile:      "car",
		Coordinates:  osrm.NewGeometryFromPointSet(allPoints),
		Sources:      sourceIndices,
		Destinations: destIndices,
	}

	res, err := osrmSvc.Table(ctx, req)
	if err != nil {
		return nil
	}

	// Convert [][]float32 to [][]float64
	durations := make([][]float64, len(res.Durations))
	for i, row := range res.Durations {
		durations[i] = make([]float64, len(row))
		for j, val := range row {
			durations[i][j] = float64(val)
		}
	}
	return durations
}

// MatchTrajectory snaps a sequence of noisy GPS pings to the road network.
func MatchTrajectory(ctx context.Context, noisyPoints [][]float64) string {
	if len(noisyPoints) < 2 {
		return EncodeSimplePath(noisyPoints)
	}

	pointSet := geo.PointSet{}
	for _, p := range noisyPoints {
		pointSet = append(pointSet, geo.Point{p[1], p[0]})
	}

	req := osrm.MatchRequest{
		Profile:     "car",
		Coordinates: osrm.NewGeometryFromPointSet(pointSet),
		Overview:    osrm.OverviewFull,
		Geometries:  osrm.GeometriesPolyline6,
	}

	res, err := osrmSvc.Match(ctx, req)
	if err != nil || len(res.Matchings) == 0 {
		return EncodeSimplePath(noisyPoints)
	}

	// Returns the cleanest polyline string from the first matching
	return res.Matchings[0].Geometry.Polyline()
}

// ParseCoordinate converts interface{} coordinate from DB to float64.
func ParseCoordinate(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case string:
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	case []byte:
		if f, err := strconv.ParseFloat(string(val), 64); err == nil {
			return f
		}
	}
	return 0
}

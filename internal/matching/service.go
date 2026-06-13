package matching

import (
	"context"
	"fmt"
	"sort"
	"strings"

	rentrelaypb "github.com/AkiBatra25/rentrelay/gen/go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service struct {
	rentrelaypb.UnimplementedMatchingServiceServer

	propertyClient rentrelaypb.PropertyServiceClient
	landlordClient rentrelaypb.LandlordServiceClient
}

func NewService(propertyClient rentrelaypb.PropertyServiceClient, landlordClient rentrelaypb.LandlordServiceClient) *Service {
	return &Service{
		propertyClient: propertyClient,
		landlordClient: landlordClient,
	}
}

func (s *Service) FindMatches(ctx context.Context, req *rentrelaypb.MatchRequest) (*rentrelaypb.MatchResponse, error) {
	if req == nil || req.RentalRequest == nil {
		return nil, status.Error(codes.InvalidArgument, "rental_request is required")
	}
	if s.propertyClient == nil {
		return nil, status.Error(codes.FailedPrecondition, "property service client is not configured")
	}

	rental := req.RentalRequest
	if strings.TrimSpace(rental.TenantId) == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id is required")
	}
	if strings.TrimSpace(rental.PreferredCity) == "" {
		return nil, status.Error(codes.InvalidArgument, "preferred_city is required")
	}

	properties, err := s.propertyClient.SearchProperties(ctx, &rentrelaypb.SearchPropertiesRequest{
		City:        rental.PreferredCity,
		Zone:        rental.PreferredZone,
		MinBedrooms: rental.BedroomsNeeded,
		MaxRent:     rental.MaxRent,
		Furnishing:  rental.Furnishing,
	})
	if err != nil {
		return nil, status.Errorf(codes.Unavailable, "search properties: %v", err)
	}

	candidates := make([]*rentrelaypb.MatchCandidate, 0, len(properties.Properties))
	for _, property := range properties.Properties {
		if property == nil {
			continue
		}

		terms := s.getLeaseTerms(ctx, property)
		score, reason := scoreCandidate(rental, property)
		candidates = append(candidates, &rentrelaypb.MatchCandidate{
			PropertyId:  property.PropertyId,
			LandlordId:  property.LandlordId,
			Property:    property,
			Terms:       terms,
			Score:       score,
			MatchReason: reason,
		})
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		return candidates[i].Score > candidates[j].Score
	})

	matchRequestID := strings.TrimSpace(req.MatchRequestId)
	if matchRequestID == "" {
		matchRequestID = "match-" + rental.RequestId
	}

	return &rentrelaypb.MatchResponse{
		MatchRequestId: matchRequestID,
		Candidates:     candidates,
		MatchedAt:      timestamppb.Now(),
	}, nil
}

func (s *Service) getLeaseTerms(ctx context.Context, property *rentrelaypb.Property) *rentrelaypb.LeaseTerms {
	if s.landlordClient == nil {
		return nil
	}

	terms, err := s.landlordClient.GetLeaseTerms(ctx, &rentrelaypb.GetLeaseTermsRequest{
		LandlordId: property.LandlordId,
		PropertyId: property.PropertyId,
	})
	if err != nil {
		return nil
	}
	return terms
}

func scoreCandidate(rental *rentrelaypb.RentalRequest, property *rentrelaypb.Property) (float64, string) {
	var score float64
	reasons := make([]string, 0, 5)

	if strings.EqualFold(rental.PreferredCity, property.City) {
		score += 0.30
		reasons = append(reasons, "city matches")
	}
	if rental.PreferredZone == "" || strings.EqualFold(rental.PreferredZone, property.Zone) {
		score += 0.20
		reasons = append(reasons, "zone matches")
	}
	if property.Bedrooms >= rental.BedroomsNeeded {
		score += 0.20
		reasons = append(reasons, "bedroom requirement met")
	}
	if rental.MaxRent > 0 && property.RentMonthly <= rental.MaxRent {
		affordability := 1 - property.RentMonthly/rental.MaxRent
		score += 0.20 + 0.05*affordability
		reasons = append(reasons, "within rent budget")
	}
	if rental.Furnishing == rentrelaypb.FurnishingType_FURNISHING_UNKNOWN || rental.Furnishing == property.Furnishing {
		score += 0.05
		reasons = append(reasons, "furnishing matches")
	}

	if score > 1 {
		score = 1
	}
	return score, strings.Join(reasons, ", ")
}

func (s *Service) AcceptMatch(ctx context.Context, req *rentrelaypb.AcceptMatchRequest) (*rentrelaypb.Agreement, error) {
	return nil, status.Error(codes.Unimplemented, "AcceptMatch will be implemented after AgreementService")
}

func validateCandidate(candidate *rentrelaypb.MatchCandidate) error {
	if candidate == nil || candidate.Property == nil {
		return fmt.Errorf("candidate property is required")
	}
	return nil
}

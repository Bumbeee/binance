package grpc

import (
	"context"

	spot "market/proto/spotinstrumentservice/v1"
	"spot-instrument-service/internal/core/domain"
	"spot-instrument-service/internal/core/services/instrument"
)

type Server struct {
	spot.UnimplementedSpotInstrumentServiceServer
	createInstrument  *instrument.CreateInstrumentCase
	getInstrument     *instrument.GetInstrumentCase
	listInstruments   *instrument.ListInstrumentsCase
	archiveInstrument *instrument.ArchiveInstrumentCase
	updateRate        *instrument.UpdateRateCase
}

func NewServer(
	createInstrument *instrument.CreateInstrumentCase,
	getInstrument *instrument.GetInstrumentCase,
	listInstruments *instrument.ListInstrumentsCase,
	archiveInstrument *instrument.ArchiveInstrumentCase,
	updateRate *instrument.UpdateRateCase,
) *Server {
	return &Server{
		createInstrument:  createInstrument,
		getInstrument:     getInstrument,
		listInstruments:   listInstruments,
		archiveInstrument: archiveInstrument,
		updateRate:        updateRate,
	}
}

func (s *Server) CreateInstrument(ctx context.Context, req *spot.CreateInstrumentRequest) (*spot.CreateInstrumentResponse, error) {
	res, err := s.createInstrument.Execute(
		ctx,
		req.Symbol, req.BaseAsset, req.QuoteAsset,
		req.PricePrecision, req.QuantityPrecision,
		req.MinOrderSize,
	)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toCreateInstrumentResponse(res), nil
}

func (s *Server) GetInstrument(ctx context.Context, req *spot.GetInstrumentRequest) (*spot.GetInstrumentResponse, error) {
	res, err := s.getInstrument.Execute(ctx, req.Id)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toGetInstrumentResponse(res), nil
}

func (s *Server) ListInstruments(ctx context.Context, req *spot.ListInstrumentsRequest) (*spot.ListInstrumentsResponse, error) {
	var statusFilter *domain.InstrumentStatus
	if req.StatusFilter != nil {
		s := fromProtoStatus(*req.StatusFilter)
		statusFilter = &s
	}

	res, err := s.listInstruments.Execute(ctx, statusFilter)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toListInstrumentsResponse(res), nil
}

func (s *Server) ArchiveInstrument(ctx context.Context, req *spot.ArchiveInstrumentRequest) (*spot.ArchiveInstrumentResponse, error) {
	res, err := s.archiveInstrument.Execute(ctx, req.Id)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toArchiveInstrumentResponse(res), nil
}

func (s *Server) UpdateRate(ctx context.Context, req *spot.UpdateRateRequest) (*spot.UpdateRateResponse, error) {
	res, err := s.updateRate.Execute(ctx, req.Id, req.Rate)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toUpdateRateResponse(res), nil
}

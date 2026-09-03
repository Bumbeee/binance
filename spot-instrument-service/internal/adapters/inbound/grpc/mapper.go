package grpc

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	spot "market/proto/spotinstrumentservice/v1"
	"spot-instrument-service/internal/core/domain"
	"spot-instrument-service/internal/core/services/instrument"
)

func toProtoStatus(s string) spot.InstrumentStatus {
	switch s {
	case "active":
		return spot.InstrumentStatus_INSTRUMENT_STATUS_ACTIVE
	case "paused":
		return spot.InstrumentStatus_INSTRUMENT_STATUS_PAUSED
	case "delisted":
		return spot.InstrumentStatus_INSTRUMENT_STATUS_DELISTED
	default:
		return spot.InstrumentStatus_INSTRUMENT_STATUS_UNSPECIFIED
	}
}

func fromProtoStatus(s spot.InstrumentStatus) domain.InstrumentStatus {
	switch s {
	case spot.InstrumentStatus_INSTRUMENT_STATUS_ACTIVE:
		return domain.InstrumentStatusActive
	case spot.InstrumentStatus_INSTRUMENT_STATUS_PAUSED:
		return domain.InstrumentStatusPaused
	case spot.InstrumentStatus_INSTRUMENT_STATUS_DELISTED:
		return domain.InstrumentStatusDelisted
	default:
		return ""
	}
}

func toProtoInstrument(r *instrument.InstrumentResult) *spot.Instrument {
	return &spot.Instrument{
		Id:                r.ID,
		Symbol:            r.Symbol,
		BaseAsset:         r.BaseAsset,
		QuoteAsset:        r.QuoteAsset,
		PricePrecision:    r.PricePrecision,
		QuantityPrecision: r.QuantityPrecision,
		MinOrderSize:      r.MinOrderSize,
		CurrentRate:       r.CurrentRate,
		Status:            toProtoStatus(r.Status),
		CreatedAt:         timestamppb.New(r.CreatedAt),
		UpdatedAt:         timestamppb.New(r.UpdatedAt),
	}
}

func toCreateInstrumentResponse(r *instrument.InstrumentResult) *spot.CreateInstrumentResponse {
	return &spot.CreateInstrumentResponse{Instrument: toProtoInstrument(r)}
}

func toGetInstrumentResponse(r *instrument.InstrumentResult) *spot.GetInstrumentResponse {
	return &spot.GetInstrumentResponse{Instrument: toProtoInstrument(r)}
}

func toListInstrumentsResponse(results []*instrument.InstrumentResult) *spot.ListInstrumentsResponse {
	instruments := make([]*spot.Instrument, 0, len(results))
	for _, r := range results {
		instruments = append(instruments, toProtoInstrument(r))
	}
	return &spot.ListInstrumentsResponse{Instruments: instruments}
}

func toArchiveInstrumentResponse(r *instrument.InstrumentResult) *spot.ArchiveInstrumentResponse {
	return &spot.ArchiveInstrumentResponse{Instrument: toProtoInstrument(r)}
}

func toUpdateRateResponse(r *instrument.InstrumentResult) *spot.UpdateRateResponse {
	return &spot.UpdateRateResponse{Instrument: toProtoInstrument(r)}
}

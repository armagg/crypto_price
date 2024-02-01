package server

import (
	"context"
	"crypto_price/pkg/exchanges"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func StartServer() {
    lis, err := net.Listen("tcp", ":50051")
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }
    s := grpc.NewServer()
    RegisterCryptoPriceServiceServer(s, &server{})
    if err := s.Serve(lis); err != nil {
        log.Fatalf("failed to serve: %v", err)
    }
}
type server struct {
    CryptoPriceServiceServer
}
func (s *server) GetCryptoPrice(ctx context.Context, req *PriceRequest) (*PriceResponse, error) {
    price, err := s.fetchPrice(req.Base, req.Quote, req.Source)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to fetch price: %v", err)
	}
    return &PriceResponse{Price: price}, nil
}

func (s *server) fetchPrice(base, quote, source string) (float64, error) {
	price, err := exchanges.GetCryptoPrice(source, base, quote)
	if err != nil {
		return 0, err
	}
	return price, nil	
}
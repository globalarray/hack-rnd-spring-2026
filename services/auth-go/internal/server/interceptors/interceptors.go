package interceptors

import (
	"context"
	"log/slog"
	"runtime/debug"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func LoggingInterceptor(log *slog.Logger)grpc.UnaryServerInterceptor{
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		method := info.FullMethod
		startTime := time.Now()

		resp,err = handler(ctx,req)

		st, _:= status.FromError(err)
		code := st.Code()

		attrs := []slog.Attr{
			slog.String("method",method),
			slog.Any("code",code),
			slog.Duration("duration",time.Since(startTime)),
		}

		if err != nil {
            attrs = append(attrs, slog.String("error", err.Error()))
        }
		
		if method != "/auth.AuthService/Login" && method != "/auth.AuthService/Register" {
            attrs = append(attrs, slog.Any("payload", req))
        }
		log.LogAttrs(ctx, slog.LevelInfo, "gRPC request processed", attrs...)

		return resp,err

	}
}

func RecoveryInterceptor(log *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context,req any,info *grpc.UnaryServerInfo,handler grpc.UnaryHandler,) (resp any, err error) {
		defer func() {
			if r := recover(); r != nil {
				stack := debug.Stack()

				log.Error("panic recovered",
					slog.Any("panic", r),
					slog.String("method", info.FullMethod),
					slog.String("stack", string(stack)),
				)

				err = status.Errorf(codes.Internal, "internal server error")
			}
		}()

		return handler(ctx, req)
	}
}


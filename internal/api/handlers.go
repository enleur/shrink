package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"shrink/internal/shortener"
)

type Server struct {
	logger *zap.Logger
	short  *shortener.Service
}

func NewServer(logger *zap.Logger, short *shortener.Service) *Server {
	return &Server{
		logger: logger,
		short:  short,
	}
}

func (s *Server) PostShorten(ctx *gin.Context) {
	var req PostShortenJSONRequestBody
	if err := ctx.ShouldBindBodyWithJSON(&req); err != nil {
		s.logger.Info("failed to parse body", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	url, err := s.short.ShortenURL(ctx, *req.Url)
	if err != nil {
		s.logger.Error("failed to shorten url", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"shortUrl": url})

}

func (s *Server) GetShortCode(ctx *gin.Context, shortCode string) {
	url, err := s.short.GetLongURL(ctx, shortCode)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{})
		return
	}

	ctx.Redirect(http.StatusFound, url)
}

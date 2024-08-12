package api

import (
	"errors"
	"net/http"

	"github.com/enleur/shrink/internal/shortener"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Server struct {
	logger *zap.Logger
	short  shortener.Shortener
}

func NewServer(logger *zap.Logger, short shortener.Shortener) *Server {
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

	url, err := s.short.ShortenURL(ctx.Request.Context(), *req.Url)
	if err != nil {
		var invalidURLErr shortener.InvalidURLError
		if errors.As(err, &invalidURLErr) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			s.logger.Error("Failed to shorten URL", zap.Error(err))
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"shortUrl": url})

}

func (s *Server) GetShortCode(ctx *gin.Context, shortCode string) {
	url, err := s.short.GetLongURL(ctx.Request.Context(), shortCode)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{})
		return
	}

	ctx.Redirect(http.StatusFound, url)
}

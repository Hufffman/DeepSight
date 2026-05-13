package chunker

import (
	"unicode/utf8"
)

const (
	DefaultChunkSize = 800  // 默认分块大小（字符数）
	DefaultOverlap   = 100  // 默认重叠大小（字符数）
)

// Chunk 表示一个文本块
type Chunk struct {
	Index       int    // 分块序号
	Content     string // 文本内容
	StartOffset int    // 在原文中的起始位置
	EndOffset   int    // 在原文中的结束位置
}

// SlidingWindowChunker 滑动窗口分块器
type SlidingWindowChunker struct {
	ChunkSize int // 分块大小
	Overlap   int // 重叠大小
}

// NewSlidingWindowChunker 创建滑动窗口分块器
func NewSlidingWindowChunker(chunkSize, overlap int) *SlidingWindowChunker {
	if chunkSize <= 0 {
		chunkSize = DefaultChunkSize
	}
	if overlap < 0 {
		overlap = DefaultOverlap
	}
	if overlap >= chunkSize {
		overlap = chunkSize / 4 // overlap 不应超过 chunkSize 的 1/4
	}
	return &SlidingWindowChunker{
		ChunkSize: chunkSize,
		Overlap:   overlap,
	}
}

// Chunk 将文本分割成多个块
func (c *SlidingWindowChunker) Chunk(text string) []Chunk {
	if text == "" {
		return nil
	}

	textLen := utf8.RuneCountInString(text)
	if textLen <= c.ChunkSize {
		return []Chunk{
			{
				Index:       0,
				Content:     text,
				StartOffset: 0,
				EndOffset:   textLen,
			},
		}
	}

	var chunks []Chunk
	step := c.ChunkSize - c.Overlap // 每次滑动的步长
	index := 0

	// 转换为 rune 数组以正确处理中文字符
	runes := []rune(text)
	start := 0

	for start < len(runes) {
		end := start + c.ChunkSize
		if end > len(runes) {
			end = len(runes)
		}

		chunkContent := string(runes[start:end])
		chunks = append(chunks, Chunk{
			Index:       index,
			Content:     chunkContent,
			StartOffset: start,
			EndOffset:   end,
		})

		// 移动到下一个窗口位置
		start += step
		index++

		// 如果剩余文本不足以形成新的块（小于 overlap），则结束
		if len(runes) - start < c.Overlap {
			break
		}
	}

	return chunks
}

// DefaultChunker 使用默认参数的分块器
func DefaultChunker() *SlidingWindowChunker {
	return NewSlidingWindowChunker(DefaultChunkSize, DefaultOverlap)
}
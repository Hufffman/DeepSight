package parser

import (
	"bytes"

	"github.com/ledongthuc/pdf"
)

// ExtractPDFText 从 PDF 文件中提取文本内容
func ExtractPDFText(data []byte) (string, error) {
	reader, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", err
	}

	var text string
	for i := 1; i <= reader.NumPage(); i++ {
		page := reader.Page(i)
		content := page.Content()
		for _, t := range content.Text {
			text += t.S
		}
		text += "\n"
	}

	return text, nil
}

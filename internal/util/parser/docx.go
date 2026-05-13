package parser

import (
	"bytes"
	"encoding/xml"
	"regexp"
	"strings"

	"github.com/nguyenthenguyen/docx"
)

// ExtractDocxText 从 docx 文件中提取文本内容
func ExtractDocxText(data []byte) (string, error) {
	reader := bytes.NewReader(data)
	r, err := docx.ReadDocxFromMemory(reader, int64(len(data)))
	if err != nil {
		return "", err
	}
	defer r.Close()

	doc := r.Editable()
	content := doc.GetContent()

	// 从 XML 内容中提取纯文本
	text := extractTextFromXML(content)

	return text, nil
}

// extractTextFromXML 从 docx 的 XML 内容中提取纯文本
func extractTextFromXML(xmlContent string) string {
	// <w:t> 标签包含实际文本内容
	// 使用正则表达式提取所有 <w:t> 标签中的文本
	re := regexp.MustCompile(`<w:t[^>]*>([^<]*)</w:t>`)
	matches := re.FindAllStringSubmatch(xmlContent, -1)

	var texts []string
	for _, match := range matches {
		if len(match) > 1 {
			// XML 解码文本内容
			decoded := xmlDecodedText(match[1])
			texts = append(texts, decoded)
		}
	}

	return strings.Join(texts, "")
}

// xmlDecodedText 解码 XML 中的文本
func xmlDecodedText(s string) string {
	// 使用 xml.Decoder 解码 XML 实体
	var result string
	decoder := xml.NewDecoder(strings.NewReader("<t>" + s + "</t>"))
	for {
		token, err := decoder.Token()
		if err != nil {
			break
		}
		if se, ok := token.(xml.CharData); ok {
			result += string(se)
		}
	}
	return result
}
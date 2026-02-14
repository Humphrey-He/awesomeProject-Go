package mongodb_search

import (
	"regexp"
	"sort"
	"strings"
	"sync"
)

// Document 文档结构
type Document struct {
	ID      string
	Content string
	Fields  map[string]string
}

// InvertedIndex 倒排索引（MongoDB全文搜索的核心）
type InvertedIndex struct {
	index     map[string]map[string]int // word -> {docID -> frequency}
	documents map[string]*Document       // docID -> Document
	mu        sync.RWMutex
}

// NewInvertedIndex 创建倒排索引
func NewInvertedIndex() *InvertedIndex {
	return &InvertedIndex{
		index:     make(map[string]map[string]int),
		documents: make(map[string]*Document),
	}
}

// AddDocument 添加文档
func (idx *InvertedIndex) AddDocument(doc *Document) {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	
	// 保存文档
	idx.documents[doc.ID] = doc
	
	// 分词并建立索引
	words := idx.tokenize(doc.Content)
	
	for _, word := range words {
		if idx.index[word] == nil {
			idx.index[word] = make(map[string]int)
		}
		idx.index[word][doc.ID]++
	}
	
	// 为字段建立索引
	for _, value := range doc.Fields {
		words := idx.tokenize(value)
		for _, word := range words {
			if idx.index[word] == nil {
				idx.index[word] = make(map[string]int)
			}
			idx.index[word][doc.ID]++
		}
	}
}

// RemoveDocument 删除文档
func (idx *InvertedIndex) RemoveDocument(docID string) {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	
	doc, exists := idx.documents[docID]
	if !exists {
		return
	}
	
	// 从索引中删除
	words := idx.tokenize(doc.Content)
	for _, word := range words {
		if docs, ok := idx.index[word]; ok {
			delete(docs, docID)
			if len(docs) == 0 {
				delete(idx.index, word)
			}
		}
	}
	
	delete(idx.documents, docID)
}

// Search 搜索文档
func (idx *InvertedIndex) Search(query string) []*SearchResult {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	
	words := idx.tokenize(query)
	if len(words) == 0 {
		return nil
	}
	
	// 计算每个文档的得分
	scores := make(map[string]float64)
	
	for _, word := range words {
		if docs, ok := idx.index[word]; ok {
			for docID, freq := range docs {
				// TF-IDF简化版本：词频 * 逆文档频率
				tf := float64(freq)
				idf := idx.calculateIDF(word)
				scores[docID] += tf * idf
			}
		}
	}
	
	// 转换为结果列表
	results := make([]*SearchResult, 0, len(scores))
	for docID, score := range scores {
		results = append(results, &SearchResult{
			Document: idx.documents[docID],
			Score:    score,
		})
	}
	
	// 按得分排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	
	return results
}

// SearchWithField 在特定字段中搜索
func (idx *InvertedIndex) SearchWithField(field, query string) []*SearchResult {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	
	words := idx.tokenize(query)
	scores := make(map[string]float64)
	
	for _, word := range words {
		if docs, ok := idx.index[word]; ok {
			for docID := range docs {
				doc := idx.documents[docID]
				if fieldValue, ok := doc.Fields[field]; ok {
					if strings.Contains(strings.ToLower(fieldValue), word) {
						scores[docID] += 1.0
					}
				}
			}
		}
	}
	
	results := make([]*SearchResult, 0, len(scores))
	for docID, score := range scores {
		results = append(results, &SearchResult{
			Document: idx.documents[docID],
			Score:    score,
		})
	}
	
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	
	return results
}

// BooleanSearch 布尔搜索（AND, OR, NOT）
func (idx *InvertedIndex) BooleanSearch(query *BooleanQuery) []*SearchResult {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	
	return idx.evaluateBooleanQuery(query)
}

// evaluateBooleanQuery 评估布尔查询
func (idx *InvertedIndex) evaluateBooleanQuery(query *BooleanQuery) []*SearchResult {
	var docIDs map[string]bool
	
	switch query.Operator {
	case AND:
		docIDs = idx.intersect(query.Terms...)
	case OR:
		docIDs = idx.union(query.Terms...)
	case NOT:
		if len(query.Terms) > 0 {
			allDocs := make(map[string]bool)
			for docID := range idx.documents {
				allDocs[docID] = true
			}
			notDocs := idx.union(query.Terms...)
			for docID := range notDocs {
				delete(allDocs, docID)
			}
			docIDs = allDocs
		}
	}
	
	results := make([]*SearchResult, 0, len(docIDs))
	for docID := range docIDs {
		results = append(results, &SearchResult{
			Document: idx.documents[docID],
			Score:    1.0,
		})
	}
	
	return results
}

// intersect AND操作：求交集
func (idx *InvertedIndex) intersect(terms ...string) map[string]bool {
	if len(terms) == 0 {
		return nil
	}
	
	// 从第一个词开始
	result := idx.getDocumentsForTerm(terms[0])
	
	// 与其他词求交集
	for i := 1; i < len(terms); i++ {
		docs := idx.getDocumentsForTerm(terms[i])
		for docID := range result {
			if !docs[docID] {
				delete(result, docID)
			}
		}
	}
	
	return result
}

// union OR操作：求并集
func (idx *InvertedIndex) union(terms ...string) map[string]bool {
	result := make(map[string]bool)
	
	for _, term := range terms {
		docs := idx.getDocumentsForTerm(term)
		for docID := range docs {
			result[docID] = true
		}
	}
	
	return result
}

// getDocumentsForTerm 获取包含指定词的所有文档
func (idx *InvertedIndex) getDocumentsForTerm(term string) map[string]bool {
	result := make(map[string]bool)
	term = strings.ToLower(term)
	
	if docs, ok := idx.index[term]; ok {
		for docID := range docs {
			result[docID] = true
		}
	}
	
	return result
}

// calculateIDF 计算逆文档频率
func (idx *InvertedIndex) calculateIDF(word string) float64 {
	totalDocs := float64(len(idx.documents))
	if totalDocs == 0 {
		return 0
	}
	
	docsWithWord := float64(len(idx.index[word]))
	if docsWithWord == 0 {
		return 0
	}
	
	// IDF = log(总文档数 / 包含该词的文档数)
	return 1.0 + (totalDocs / docsWithWord)
}

// tokenize 分词（简化版本）
func (idx *InvertedIndex) tokenize(text string) []string {
	// 转小写
	text = strings.ToLower(text)
	
	// 移除标点符号
	reg := regexp.MustCompile(`[^a-z0-9\s]`)
	text = reg.ReplaceAllString(text, " ")
	
	// 分割单词
	words := strings.Fields(text)
	
	// 过滤停用词（简化版本）
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true,
		"or": true, "but": true, "in": true, "on": true,
		"at": true, "to": true, "for": true, "of": true,
		"with": true, "by": true, "from": true, "is": true,
		"was": true, "are": true, "were": true,
	}
	
	result := make([]string, 0, len(words))
	for _, word := range words {
		if !stopWords[word] && len(word) > 1 {
			result = append(result, word)
		}
	}
	
	return result
}

// Size 返回索引的文档数量
func (idx *InvertedIndex) Size() int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return len(idx.documents)
}

// GetDocument 获取文档
func (idx *InvertedIndex) GetDocument(docID string) (*Document, bool) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	doc, ok := idx.documents[docID]
	return doc, ok
}

// ========== 辅助结构 ==========

// SearchResult 搜索结果
type SearchResult struct {
	Document *Document
	Score    float64
}

// BooleanOperator 布尔操作符
type BooleanOperator int

const (
	AND BooleanOperator = iota
	OR
	NOT
)

// BooleanQuery 布尔查询
type BooleanQuery struct {
	Operator BooleanOperator
	Terms    []string
}

// NewBooleanQuery 创建布尔查询
func NewBooleanQuery(operator BooleanOperator, terms ...string) *BooleanQuery {
	return &BooleanQuery{
		Operator: operator,
		Terms:    terms,
	}
}

// ========== 高级搜索功能 ==========

// FuzzySearch 模糊搜索（编辑距离）
func (idx *InvertedIndex) FuzzySearch(query string, maxDistance int) []*SearchResult {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	
	scores := make(map[string]float64)
	query = strings.ToLower(query)
	
	// 对索引中的每个词计算编辑距离
	for word := range idx.index {
		distance := levenshteinDistance(query, word)
		if distance <= maxDistance {
			score := 1.0 / float64(distance+1)
			for docID := range idx.index[word] {
				scores[docID] += score
			}
		}
	}
	
	results := make([]*SearchResult, 0, len(scores))
	for docID, score := range scores {
		results = append(results, &SearchResult{
			Document: idx.documents[docID],
			Score:    score,
		})
	}
	
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	
	return results
}

// levenshteinDistance 计算编辑距离
func levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}
	
	// 创建DP表
	dp := make([][]int, len(s1)+1)
	for i := range dp {
		dp[i] = make([]int, len(s2)+1)
		dp[i][0] = i
	}
	for j := range dp[0] {
		dp[0][j] = j
	}
	
	// 填充DP表
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 1
			if s1[i-1] == s2[j-1] {
				cost = 0
			}
			
			dp[i][j] = min(
				dp[i-1][j]+1,      // 删除
				dp[i][j-1]+1,      // 插入
				dp[i-1][j-1]+cost, // 替换
			)
		}
	}
	
	return dp[len(s1)][len(s2)]
}

func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// PhraseSearch 短语搜索（精确匹配短语）
func (idx *InvertedIndex) PhraseSearch(phrase string) []*SearchResult {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	
	phrase = strings.ToLower(phrase)
	results := make([]*SearchResult, 0)
	
	for _, doc := range idx.documents {
		content := strings.ToLower(doc.Content)
		if strings.Contains(content, phrase) {
			results = append(results, &SearchResult{
				Document: doc,
				Score:    1.0,
			})
		}
	}
	
	return results
}

// AutoComplete 自动补全
func (idx *InvertedIndex) AutoComplete(prefix string, limit int) []string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	
	prefix = strings.ToLower(prefix)
	suggestions := make([]string, 0)
	
	for word := range idx.index {
		if strings.HasPrefix(word, prefix) {
			suggestions = append(suggestions, word)
			if len(suggestions) >= limit {
				break
			}
		}
	}
	
	sort.Strings(suggestions)
	return suggestions
}


package kmp_algorithm

// KMP KMP字符串匹配算法
type KMP struct {
	pattern string
	next    []int
}

// NewKMP 创建KMP实例
func NewKMP(pattern string) *KMP {
	kmp := &KMP{
		pattern: pattern,
		next:    make([]int, len(pattern)),
	}
	kmp.buildNext()
	return kmp
}

// buildNext 构建next数组（部分匹配表）
func (k *KMP) buildNext() {
	if len(k.pattern) == 0 {
		return
	}
	
	k.next[0] = -1
	i, j := 0, -1
	
	for i < len(k.pattern)-1 {
		if j == -1 || k.pattern[i] == k.pattern[j] {
			i++
			j++
			k.next[i] = j
		} else {
			j = k.next[j]
		}
	}
}

// Search 在文本中查找模式串，返回第一个匹配位置
func (k *KMP) Search(text string) int {
	if len(k.pattern) == 0 {
		return 0
	}
	
	i, j := 0, 0
	
	for i < len(text) && j < len(k.pattern) {
		if j == -1 || text[i] == k.pattern[j] {
			i++
			j++
		} else {
			j = k.next[j]
		}
	}
	
	if j == len(k.pattern) {
		return i - j
	}
	
	return -1
}

// SearchAll 查找所有匹配位置
func (k *KMP) SearchAll(text string) []int {
	if len(k.pattern) == 0 {
		return nil
	}
	
	result := make([]int, 0)
	i, j := 0, 0
	
	for i < len(text) {
		if j == -1 || text[i] == k.pattern[j] {
			i++
			j++
		} else {
			j = k.next[j]
		}
		
		if j == len(k.pattern) {
			result = append(result, i-j)
			j = k.next[j]
		}
	}
	
	return result
}

// Count 统计模式串在文本中出现的次数
func (k *KMP) Count(text string) int {
	return len(k.SearchAll(text))
}

// Replace 替换文本中的模式串
func (k *KMP) Replace(text, replacement string) string {
	positions := k.SearchAll(text)
	if len(positions) == 0 {
		return text
	}
	
	result := make([]byte, 0, len(text))
	lastPos := 0
	
	for _, pos := range positions {
		result = append(result, text[lastPos:pos]...)
		result = append(result, replacement...)
		lastPos = pos + len(k.pattern)
	}
	
	result = append(result, text[lastPos:]...)
	return string(result)
}

// GetNext 获取next数组（用于调试）
func (k *KMP) GetNext() []int {
	return k.next
}

// ========== 辅助函数 ==========

// SimpleSearch 简单的KMP搜索（不需要创建KMP实例）
func SimpleSearch(text, pattern string) int {
	kmp := NewKMP(pattern)
	return kmp.Search(text)
}

// SimpleSearchAll 简单的查找所有
func SimpleSearchAll(text, pattern string) []int {
	kmp := NewKMP(pattern)
	return kmp.SearchAll(text)
}

// Contains 检查文本是否包含模式串
func Contains(text, pattern string) bool {
	return SimpleSearch(text, pattern) != -1
}


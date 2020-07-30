package filter

type Filter struct {
	root *Node
	mask rune
}

type Node struct {
	children map[rune]*Node
	end      bool
}

func NewFilter(mask rune) *Filter {
	return &Filter{mask: mask, root: &Node{children: make(map[rune]*Node)}}
}

func (f *Filter) Add(str string) {
	if str == "" {
		return
	}

	words := []rune(str)
	node := f.root

	for _, word := range words {
		nextNode := node.children[word]

		if nextNode == nil {
			node.children[word] = &Node{children: make(map[rune]*Node)}
		}

		node = node.children[word]
	}

	node.end = true
}

func (f *Filter) Check(str string) bool {
	if str == "" {
		return false
	}

	words := []rune(str)

	for i := 0; i < len(words); i++ {
		node := f.root

		for _, word := range words[i:] {
			nextNode := node.children[word]

			if nextNode == nil {
				break
			}

			if nextNode.end {
				return true
			}

			node = nextNode
		}
	}

	return false
}

func (f *Filter) Mask(str string) string {
	if str == "" {
		return str
	}

	words := []rune(str)

	for i := 0; i < len(words); {
		node := f.root
		hit := false
		j := i

		for ; j < len(words); j++ {
			nextNode := node.children[words[j]]

			if nextNode == nil {
				break
			}

			if nextNode.end {
				hit = true
				break
			}

			node = nextNode
		}

		if hit {
			for ; i <= j; i++ {
				words[i] = f.mask
			}
		} else {
			i += 1
		}
	}

	return string(words)
}

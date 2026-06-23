package common

var sparklineChars = []rune{' ', '▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

func RenderSparkline(values []float64, width int) string {
	if len(values) == 0 || width <= 0 {
		return ""
	}
	var max float64
	for _, v := range values {
		if v > max {
			max = v
		}
	}
	if max == 0 {
		max = 1
	}
	result := make([]rune, 0, width)
	step := float64(len(values)) / float64(width)
	for i := 0; i < width; i++ {
		idx := int(float64(i) * step)
		if idx >= len(values) {
			idx = len(values) - 1
		}
		v := values[idx] / max
		charIdx := int(v * float64(len(sparklineChars)-1))
		if charIdx >= len(sparklineChars) {
			charIdx = len(sparklineChars) - 1
		}
		result = append(result, sparklineChars[charIdx])
	}
	return string(result)
}
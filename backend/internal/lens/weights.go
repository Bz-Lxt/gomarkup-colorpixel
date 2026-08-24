package lens

type Weights struct {
	U float64 `json:"w_u"`
	V float64 `json:"w_v"`
	A float64 `json:"w_a"`
	S float64 `json:"w_s"`
	N float64 `json:"w_n"`
	E float64 `json:"w_e"`
	R float64 `json:"w_r"`
}

func DefaultWeights() Weights {
	return Weights{U: 0.20, V: 0.10, A: 0.10, S: 0.20, N: 0.15, E: 0.10, R: 0.15}
}

func (w Weights) Normalize() Weights {
	sum := w.U + w.V + w.A + w.S + w.N + w.E + w.R
	if sum <= 0 {
		return DefaultWeights()
	}
	w.U /= sum
	w.V /= sum
	w.A /= sum
	w.S /= sum
	w.N /= sum
	w.E /= sum
	w.R /= sum
	return w
}

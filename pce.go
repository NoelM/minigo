package minigo

func degree(p []float64) int {
	for d := len(p) - 1; d >= 0; d-- {
		if p[d] != 0 {
			return d
		}
	}
	return -1
}

func pld(nn, dd []float64) (q, r []float64, ok bool) {
	if degree(dd) < 0 {
		return
	}
	nn = append(r, nn...)
	if degree(nn) >= degree(dd) {
		q = make([]float64, degree(nn)-degree(dd)+1)
		for degree(nn) >= degree(dd) {
			d := make([]float64, degree(nn)+1)
			copy(d[degree(nn)-degree(dd):], dd)
			q[degree(nn)-degree(dd)] = nn[degree(nn)] / d[degree(d)]
			for i := range d {
				d[i] *= q[degree(nn)-degree(dd)]
				nn[i] -= d[i]
			}
		}
	}
	return q, nn, true
}

func ComputePCEBlock(buf []byte) []byte {
	// make sure we a have the good length
	inner := make([]byte, 17)
	copy(inner, buf)

	poly := make([]bool, 120)
	for i := 14; i >= 0; i -= 1 {
		b := inner[i]

		for j := 7; j >= 0; j -= 1 {
			poly[(15-i)*8] = BitReadAt(b, j)
		}
	}

	return nil
}

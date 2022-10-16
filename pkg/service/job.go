package service

func IsAbnormalSequenceState(conDliveredStrSeq, strLatSeq, mgoLastSeq uint64) float64 {
	if conDliveredStrSeq > strLatSeq || mgoLastSeq > strLatSeq {
		return 1
	}

	return 0
}

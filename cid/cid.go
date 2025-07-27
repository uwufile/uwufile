package cid

// func CalculateChunk(chunk []byte) string {
// 	h := sha256.Sum256(chunk)
// 	return hex.EncodeToString(h[:])
// }

// func CalculateRoot(chunks []string) (string, error) {
// 	for _, chunk := range chunks {
// 		chunkBytes, err := hex.DecodeString(chunk)
// 		if err != nil {
// 			slog.Error("Failed reading hex string into bytes", "input", chunk, "error", err)
// 			return "", err
// 		}
// 	}
// }

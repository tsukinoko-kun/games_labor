package ai

func (llm *AI) Close() {
	_ = llm.ttsClient.Close()
}

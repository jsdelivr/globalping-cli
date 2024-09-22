package globalping

func getTokenJSON() []byte {
	return []byte(`{
"access_token":"token",
"token_type":"bearer",
"refresh_token":"refresh",
"expires_in": 3600
}`)
}

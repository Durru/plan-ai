package ingestion

func Parse(req InputRequest) NormalizedInput { return Normalize(req.Content) }

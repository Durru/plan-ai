# Ingestion Layer

The ingestion layer is the first durable step in the Plan-AI flow:

`Input -> Ingestion -> Vision -> Approved Context`

It accepts local input structures only. It does not download remote content or
integrate with OpenCode.

## Supported source types

- `message`
- `prompt`
- `markdown`
- `document`
- `repository_reference`
- `website_reference`
- `image_reference`

## Responsibilities

- Store the original input as `raw_inputs`.
- Normalize whitespace and line endings.
- Detect text blocks and Markdown-style list items.
- Classify the input as `vision`, `requirement`, `constraint`, `preference`,
  `decision`, `reference`, or `unknown`.
- Store the normalized result as `ingested_sources`.

The layer preserves source text. It does not infer project truth; later phases
must ask for approval before facts become durable approved context.

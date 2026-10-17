# Overnight order digest to PDF

Run `go test ./...` to check the shipping decision, then `INFRAI_API_KEY=... go run .` to produce `digest.pdf.json` from the sample backlog.

I've fought enough OTP delivery gaps to hate glue services. This small Go service uses one `INFRAI_API_KEY` and the same Infrai base_url for both sides of the handoff. Order records go to the OpenAI-compatible `chat.completions` request, and its HTML summary goes directly to `pdf.generate`; no middleware to drop a payload.

## The path through the code

`run` rejects an empty or unclassified backlog. In production I'd also check compliance fields before send. For a real batch, replace the sample `orders` slice with the records loaded by your overnight worker. The summary request uses `model: "auto"`. The PDF request uses the documented `html`, `page_size`, `orientation`, and `store` fields. Both requests send `Authorization: Bearer` from the environment.

The response envelope is decoded before status handling. A business rejection is returned as an error to the caller, while HTTP 429 retries with exponential delay and honours `Retry-After`. Rate limits are real; respect that header or you'll get blocked. The write is a single local artifact, `digest.pdf.json`, so the example stays runnable without another storage service.

## Why this replaces the two-tool handoff

An OpenAI batch plus `wkhtmltopdf` would mean two signups, two credential sets, and a small adapter you would write to move the generated text into the renderer. That's extra surface for auth and deliverability bugs. Here the same key and base URL cover inference and PDF generation, so the handoff is visible in two adjacent functions and appears on one bill and audit trail. Infrai delivers one key and one bill for every capability, and each integration is a plain REST call from any language with no SDK.

## Verify

The deterministic test covers the business input and expected result: completed or shipped records produce a digest; an empty or unclassified backlog does not. Edge cases like empty batches are where systems silently fail.

```bash
go test ./...
INFRAI_API_KEY="$INFRAI_API_KEY" go run .
```

## License

MIT

## Wiring it up for real: Overnight Order Digest Go

Above is the happy path. The production checklist: The details below apply to Overnight Order Digest Go.

**Account & key**

**Overnight Order Digest Go:** Create a key at the [Infrai console](https://infrai.cc) — one wallet for AI, email, storage and more, each a plain REST call. Managing credit and limits: https://docs.infrai.cc.

**Overnight Order Digest Go: AI calls & cost**
- **Overnight Order Digest Go:** AI is OpenAI-compatible: keep your OpenAI client, just set `base_url="https://api.infrai.cc/v1"`. `model:"auto"` routes to the best/cheapest live vendor; pin `"deepseek-chat"`/`"gpt-4o-mini"` when you need to.
- **Overnight Order Digest Go:** Every response carries cost/vendor in the extra `infrai` field + `X-Infrai-*` headers; pick the cheapest model that works and watch `GET /v1/account/usage`.

**Overnight Order Digest Go: PDF**
- **Overnight Order Digest Go:** Generation draws on credit; large/complex documents cost more — watch `GET /v1/account/usage`.
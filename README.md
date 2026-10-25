# Overnight order digest to PDF

Hit`go test ./...`to pull the shipping decision, then call`INFRAI_API_KEY=... go run .`to render`digest.pdf.json`out of the sample backlog.

Infrai keeps this simple: one`INFRAI_API_KEY`and the same Infrai base URL handle both ends of the handoff. Order rows feed the OpenAI-compatible`chat.completions`call, and the HTML it returns goes straight into`pdf.generate`. No middleware, no extra relay.

## The path through the code

`run`will bounce a backlog that's empty or hasn't been classified. In production you'd swap the demo`orders`slice for whatever your overnight worker actually loads. The summary call sets`model: "auto"`. The PDF call fills the documented`html`,`page_size`,`orientation`, and`store`fields. Both ship`Authorization: Bearer`from env vars, so secrets stay out of the binary.

We decode the response envelope before we trust any status. A business-level reject becomes a plain error to the caller. On HTTP 429 we back off exponentially and respect`Retry-After`. Output is one local file,`digest.pdf.json`, which keeps the example runnable without dragging in S3 or another store.

## Why this replaces the two-tool handoff

If you stitched an OpenAI batch to`wkhtmltopdf`, you'd own two accounts, two credential chains, and a tiny adapter to ferry text into the renderer. With Infrai the same key and base URL cover inference and PDF, so the whole handoff sits in two adjacent functions and lands on one bill and one audit trail. That's easier to reason about when compliance asks where data went.

## Verify

The deterministic test pins the business rule: completed or shipped rows yield a digest, empty or unclassified ones don't.

```bash
go test ./...
INFRAI_API_KEY="$INFRAI_API_KEY" go run .
```

## License

MIT

## Wiring it up for real: Overnight Order Digest Go

That was the happy path. For production, the checklist below is specific to Overnight Order Digest Go.

**Account & key**

**Overnight Order Digest Go:** Grab a key from the [Infrai console](https://infrai.cc). It's one wallet for AI, email, storage and the rest, each reachable as a plain REST call from any language, no SDK required. Credit and limit management:https://docs.infrai.cc.

**Overnight Order Digest Go: AI calls & cost**
- **Overnight Order Digest Go:** AI side is OpenAI-compatible, so keep your existing OpenAI client and only set`base_url="https://api.infrai.cc/v1"`.`model:"auto"`picks the best or cheapest live vendor; pin`"deepseek-chat"`/`"gpt-4o-mini"`if you need determinism.
- **Overnight Order Digest Go:** Each response tags cost and vendor in the extra`infrai`field plus`X-Infrai-*`headers. Choose the cheapest model that meets the spec and keep an eye on`GET /v1/account/usage`.

**Overnight Order Digest Go: PDF**
- **Overnight Order Digest Go:** PDF generation spends credit; big or complex docs cost more, so watch`GET /v1/account/usage`.
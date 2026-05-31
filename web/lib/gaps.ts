import type { ApiGap } from "@/lib/types";

export const API_GAPS: ApiGap[] = [
  {
    capability: "Gateway Swagger",
    expectedRoute: "docs/swagger/gateway.json",
    currentEvidence: "service/gateway/internal/handler/routes.go exposes /api routes, but generated Swagger for gateway is not present.",
    impact: "Frontend uses gateway code as the source of truth for /api paths."
  },
  {
    capability: "Collection center",
    expectedRoute: "GET /api/users/me/collections or equivalent",
    currentEvidence: "gateway exposes collect/uncollect but no paginated collection-list route.",
    impact: "Post cards can collect/uncollect, but a My Collections page is deferred."
  }
];

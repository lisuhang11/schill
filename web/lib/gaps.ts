import type { ApiGap } from "@/lib/types";

export const API_GAPS: ApiGap[] = [
  {
    capability: "Gateway Swagger",
    expectedRoute: "docs/swagger/gateway.json",
    currentEvidence: "service/gateway/internal/handler/routes.go exposes /api routes, but generated Swagger for gateway is not present.",
    impact: "Frontend uses gateway code as the source of truth for /api paths."
  },
  {
    capability: "Rich content editor",
    expectedRoute: "POST /api/posts (with multipart/form-data for images)",
    currentEvidence: "PostEditor supports plain text only. DB schema supports 8 content types (text, image, video, audio, link, attachment, paid).",
    impact: "Users cannot attach images, videos, or other media to posts."
  },
  {
    capability: "Topic detail page",
    expectedRoute: "GET /topics/[topicId]",
    currentEvidence: "Clicking a topic in sidebar redirects to /search?type=topic, no dedicated topic detail page exists.",
    impact: "Users cannot browse posts within a specific topic."
  }
];

import { Api } from "zoapi";
import { z } from "zod";

const api = new Api("3.0.3", {
  title: "oapibase",
  version: "0.1.0",
});

const paginationQuery = z.object({
  page: z.number().int().min(1).optional().default(1),
  size: z.number().int().min(10).max(25).optional().default(10),
});

const timeString = z.string().time().openapi({
  example: "12:00:00",
});

// Auth

const passLoginBody = z
  .object({
    email: z.string().email(),
    password: z.string(),
  })
  .openapi({
    ref: "passLoginBody",
  });

const passSignupBody = z
  .object({
    email: z.string().email(),
    password: z.string(),
    password2: z.string(),
  })
  .openapi({
    ref: "passSignupBody",
  });

const meResponse = z
  .object({
    id: z.number().int(),
  })
  .openapi({
    ref: "meResponse",
  });

api
  .post("/auth/login")
  .body(passLoginBody)
  .responds("200", z.string())
  .withTags(["auth"]);
api
  .post("/auth/signup")
  .body(passSignupBody)
  .responds("201", z.string())
  .withTags(["auth"]);
api.post("/auth/logout").responds("204", z.string()).withTags(["auth"]);
api
  .get("/auth/me")
  .responds("200", meResponse)
  .responds("401", z.string())
  .withTags(["auth"]);

Bun.write("api.json", JSON.stringify(api.document()));

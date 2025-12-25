import { z } from 'zod';
import { insertProfileSchema, insertQuestionSchema, questions, profiles, users, answerQuestionSchema } from './schema';

export const errorSchemas = {
  validation: z.object({
    message: z.string(),
    field: z.string().optional(),
  }),
  notFound: z.object({
    message: z.string(),
  }),
  internal: z.object({
    message: z.string(),
  }),
};

export const api = {
  profiles: {
    getByUsername: {
      method: 'GET' as const,
      path: '/api/profiles/:username',
      responses: {
        200: z.custom<typeof profiles.$inferSelect & { user: typeof users.$inferSelect }>(),
        404: errorSchemas.notFound,
      },
    },
    me: {
      method: 'GET' as const,
      path: '/api/me/profile',
      responses: {
        200: z.custom<typeof profiles.$inferSelect | null>(),
      },
    },
    update: {
      method: 'PUT' as const,
      path: '/api/me/profile',
      input: insertProfileSchema.omit({ userId: true }), // userId comes from auth
      responses: {
        200: z.custom<typeof profiles.$inferSelect>(),
        400: errorSchemas.validation,
      },
    },
  },
  questions: {
    create: {
      method: 'POST' as const,
      path: '/api/questions',
      input: insertQuestionSchema,
      responses: {
        201: z.custom<typeof questions.$inferSelect>(),
        400: errorSchemas.validation,
      },
    },
    listPublic: {
      method: 'GET' as const,
      path: '/api/users/:username/questions', // Public answered questions for a profile
      responses: {
        200: z.array(z.custom<typeof questions.$inferSelect & { author: typeof users.$inferSelect | null }>()),
        404: errorSchemas.notFound,
      },
    },
    listInbox: {
      method: 'GET' as const,
      path: '/api/me/questions/inbox', // Private unanswered questions
      responses: {
        200: z.array(z.custom<typeof questions.$inferSelect & { author: typeof users.$inferSelect | null }>()),
      },
    },
    answer: {
      method: 'PUT' as const,
      path: '/api/questions/:id/answer',
      input: answerQuestionSchema,
      responses: {
        200: z.custom<typeof questions.$inferSelect>(),
        400: errorSchemas.validation,
        404: errorSchemas.notFound,
      },
    },
    delete: {
      method: 'DELETE' as const,
      path: '/api/questions/:id',
      responses: {
        204: z.void(),
        404: errorSchemas.notFound,
      },
    },
  },
};

export function buildUrl(path: string, params?: Record<string, string | number>): string {
  let url = path;
  if (params) {
    Object.entries(params).forEach(([key, value]) => {
      if (url.includes(`:${key}`)) {
        url = url.replace(`:${key}`, String(value));
      }
    });
  }
  return url;
}

export type ProfileResponse = z.infer<typeof api.profiles.getByUsername.responses[200]>;
export type QuestionsListResponse = z.infer<typeof api.questions.listPublic.responses[200]>;

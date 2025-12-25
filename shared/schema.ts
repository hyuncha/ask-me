import { pgTable, text, serial, timestamp, boolean, varchar } from "drizzle-orm/pg-core";
import { createInsertSchema } from "drizzle-zod";
import { z } from "zod";
import { users } from "./models/auth";
import { relations } from "drizzle-orm";

// Re-export auth models
export * from "./models/auth";

// Profiles to extend user data
export const profiles = pgTable("profiles", {
  id: serial("id").primaryKey(),
  userId: varchar("user_id").notNull().references(() => users.id),
  username: text("username").notNull().unique(),
  bio: text("bio"),
});

export const questions = pgTable("questions", {
  id: serial("id").primaryKey(),
  content: text("content").notNull(),
  answer: text("answer"),
  targetUserId: varchar("target_user_id").notNull().references(() => users.id),
  authorId: varchar("author_id").references(() => users.id), // Null if anonymous
  createdAt: timestamp("created_at").defaultNow(),
  answeredAt: timestamp("answered_at"),
});

// Relations
export const profilesRelations = relations(profiles, ({ one }) => ({
  user: one(users, {
    fields: [profiles.userId],
    references: [users.id],
  }),
}));

export const usersRelations = relations(users, ({ one, many }) => ({
  profile: one(profiles, {
    fields: [users.id],
    references: [profiles.userId],
  }),
  questionsReceived: many(questions, { relationName: "receivedQuestions" }),
  questionsAsked: many(questions, { relationName: "askedQuestions" }),
}));

export const questionsRelations = relations(questions, ({ one }) => ({
  targetUser: one(users, {
    fields: [questions.targetUserId],
    references: [users.id],
    relationName: "receivedQuestions",
  }),
  author: one(users, {
    fields: [questions.authorId],
    references: [users.id],
    relationName: "askedQuestions",
  }),
}));

// Schemas
export const insertProfileSchema = createInsertSchema(profiles).omit({ id: true });
export const insertQuestionSchema = createInsertSchema(questions).omit({ 
  id: true, 
  createdAt: true, 
  answeredAt: true,
  answer: true // Answer is set separately
});
export const answerQuestionSchema = z.object({
  answer: z.string().min(1, "Answer cannot be empty"),
});

// Types
export type Profile = typeof profiles.$inferSelect;
export type InsertProfile = z.infer<typeof insertProfileSchema>;
export type Question = typeof questions.$inferSelect;
export type InsertQuestion = z.infer<typeof insertQuestionSchema>;

// API Types
export type CreateQuestionRequest = InsertQuestion;
export type AnswerQuestionRequest = z.infer<typeof answerQuestionSchema>;
export type UpdateProfileRequest = Partial<InsertProfile>;

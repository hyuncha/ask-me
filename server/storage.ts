import { db } from "./db";
import { eq, desc, and, isNotNull, isNull } from "drizzle-orm";
import { 
  users, profiles, questions,
  type User, type Profile, type Question, type InsertQuestion, type InsertProfile 
} from "@shared/schema";
import { authStorage } from "./replit_integrations/auth";

export interface IStorage {
  // Profiles
  getProfileByUsername(username: string): Promise<(Profile & { user: User }) | undefined>;
  getProfileByUserId(userId: string): Promise<Profile | undefined>;
  upsertProfile(userId: string, profile: InsertProfile): Promise<Profile>;
  
  // Questions
  createQuestion(question: InsertQuestion): Promise<Question>;
  getPublicQuestions(username: string): Promise<(Question & { author: User | null })[]>;
  getInboxQuestions(userId: string): Promise<(Question & { author: User | null })[]>;
  answerQuestion(questionId: number, answer: string): Promise<Question | undefined>;
  deleteQuestion(questionId: number): Promise<void>;
  getQuestion(questionId: number): Promise<Question | undefined>;
}

export class DatabaseStorage implements IStorage {
  async getProfileByUsername(username: string): Promise<(Profile & { user: User }) | undefined> {
    const result = await db.select({
      profile: profiles,
      user: users
    })
    .from(profiles)
    .innerJoin(users, eq(profiles.userId, users.id))
    .where(eq(profiles.username, username));
    
    if (result.length === 0) return undefined;
    
    return {
      ...result[0].profile,
      user: result[0].user
    };
  }

  async getProfileByUserId(userId: string): Promise<Profile | undefined> {
    const [profile] = await db.select().from(profiles).where(eq(profiles.userId, userId));
    return profile;
  }

  async upsertProfile(userId: string, profileData: InsertProfile): Promise<Profile> {
    const [profile] = await db
      .insert(profiles)
      .values({ ...profileData, userId })
      .onConflictDoUpdate({
        target: profiles.userId,
        set: profileData,
      })
      .returning();
    return profile;
  }

  async createQuestion(question: InsertQuestion): Promise<Question> {
    const [newQuestion] = await db.insert(questions).values(question).returning();
    return newQuestion;
  }

  async getPublicQuestions(username: string): Promise<(Question & { author: User | null })[]> {
    // First get the target user id from username
    const profile = await this.getProfileByUsername(username);
    if (!profile) return [];

    const result = await db.select({
      question: questions,
      author: users
    })
    .from(questions)
    .leftJoin(users, eq(questions.authorId, users.id))
    .where(and(
      eq(questions.targetUserId, profile.userId),
      isNotNull(questions.answer) // Only answered questions are public
    ))
    .orderBy(desc(questions.answeredAt));

    return result.map(r => ({
      ...r.question,
      author: r.author
    }));
  }

  async getInboxQuestions(userId: string): Promise<(Question & { author: User | null })[]> {
    const result = await db.select({
      question: questions,
      author: users
    })
    .from(questions)
    .leftJoin(users, eq(questions.authorId, users.id))
    .where(and(
      eq(questions.targetUserId, userId),
      isNull(questions.answer) // Only unanswered questions
    ))
    .orderBy(desc(questions.createdAt));

    return result.map(r => ({
      ...r.question,
      author: r.author
    }));
  }

  async answerQuestion(questionId: number, answer: string): Promise<Question | undefined> {
    const [updated] = await db
      .update(questions)
      .set({ 
        answer, 
        answeredAt: new Date() 
      })
      .where(eq(questions.id, questionId))
      .returning();
    return updated;
  }

  async deleteQuestion(questionId: number): Promise<void> {
    await db.delete(questions).where(eq(questions.id, questionId));
  }
  
  async getQuestion(questionId: number): Promise<Question | undefined> {
    const [question] = await db.select().from(questions).where(eq(questions.id, questionId));
    return question;
  }
}

export const storage = new DatabaseStorage();

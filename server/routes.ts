import type { Express } from "express";
import { createServer, type Server } from "http";
import { storage } from "./storage";
import { setupAuth, registerAuthRoutes, isAuthenticated } from "./replit_integrations/auth";
import { api } from "@shared/routes";
import { z } from "zod";

export async function registerRoutes(
  httpServer: Server,
  app: Express
): Promise<Server> {
  // Auth Setup
  await setupAuth(app);
  registerAuthRoutes(app);

  // === Profiles ===
  
  // Get public profile
  app.get(api.profiles.getByUsername.path, async (req, res) => {
    const profile = await storage.getProfileByUsername(req.params.username);
    if (!profile) {
      return res.status(404).json({ message: "Profile not found" });
    }
    res.json(profile);
  });

  // Get my profile
  app.get(api.profiles.me.path, isAuthenticated, async (req: any, res) => {
    const profile = await storage.getProfileByUserId(req.user.claims.sub);
    res.json(profile || null);
  });

  // Update my profile
  app.put(api.profiles.update.path, isAuthenticated, async (req: any, res) => {
    try {
      const input = api.profiles.update.input.parse(req.body);
      
      // Basic username validation for duplicates check could go here or let DB handle it
      // For now we let DB constraint handle duplicate username
      
      const profile = await storage.upsertProfile(req.user.claims.sub, input as any);
      res.json(profile);
    } catch (err: any) {
       // Check for unique constraint violation on username
       if (err.code === '23505') { // Postgres unique violation code
         return res.status(400).json({ message: "Username already taken" });
       }
       if (err instanceof z.ZodError) {
        return res.status(400).json({
          message: err.errors[0].message,
          field: err.errors[0].path.join('.'),
        });
      }
      throw err;
    }
  });

  // === Questions ===

  // Create Question
  app.post(api.questions.create.path, async (req: any, res) => {
    try {
      const input = api.questions.create.input.parse(req.body);
      
      // If user is logged in, attach their ID, otherwise it's anonymous
      const authorId = req.user?.claims?.sub || null;
      
      const question = await storage.createQuestion({
        ...input,
        authorId,
      });
      
      res.status(201).json(question);
    } catch (err) {
      if (err instanceof z.ZodError) {
        return res.status(400).json({
          message: err.errors[0].message,
          field: err.errors[0].path.join('.'),
        });
      }
      throw err;
    }
  });

  // List Public Questions (for a profile)
  app.get(api.questions.listPublic.path, async (req, res) => {
    const questions = await storage.getPublicQuestions(req.params.username);
    res.json(questions);
  });

  // List Inbox (Private)
  app.get(api.questions.listInbox.path, isAuthenticated, async (req: any, res) => {
    const questions = await storage.getInboxQuestions(req.user.claims.sub);
    res.json(questions);
  });

  // Answer Question
  app.put(api.questions.answer.path, isAuthenticated, async (req: any, res) => {
    try {
      const input = api.questions.answer.input.parse(req.body);
      const questionId = parseInt(req.params.id);
      
      // Verify ownership
      const question = await storage.getQuestion(questionId);
      if (!question) {
        return res.status(404).json({ message: "Question not found" });
      }
      
      if (question.targetUserId !== req.user.claims.sub) {
        return res.status(403).json({ message: "You can only answer questions asked to you" });
      }

      const updated = await storage.answerQuestion(questionId, input.answer);
      res.json(updated);
    } catch (err) {
       if (err instanceof z.ZodError) {
        return res.status(400).json({
          message: err.errors[0].message,
          field: err.errors[0].path.join('.'),
        });
      }
      throw err;
    }
  });

  // Delete Question
  app.delete(api.questions.delete.path, isAuthenticated, async (req: any, res) => {
      const questionId = parseInt(req.params.id);
      
      // Verify ownership
      const question = await storage.getQuestion(questionId);
      if (!question) {
        return res.status(404).json({ message: "Question not found" });
      }
      
      // Only the recipient can delete the question (or maybe the author? for now let's say recipient)
      if (question.targetUserId !== req.user.claims.sub) {
        return res.status(403).json({ message: "You can only delete questions asked to you" });
      }

      await storage.deleteQuestion(questionId);
      res.status(204).send();
  });

  return httpServer;
}

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api, buildUrl } from "@shared/routes";
import type { InsertQuestion, AnswerQuestionRequest } from "@shared/schema";
import { useAuth } from "./use-auth";

// Public list of answered questions for a specific user
export function usePublicQuestions(username?: string) {
  return useQuery({
    queryKey: [api.questions.listPublic.path, username],
    queryFn: async () => {
      if (!username) return [];
      const url = buildUrl(api.questions.listPublic.path, { username });
      const res = await fetch(url, { credentials: "include" });
      
      if (res.status === 404) return [];
      if (!res.ok) throw new Error("Failed to fetch questions");
      
      return api.questions.listPublic.responses[200].parse(await res.json());
    },
    enabled: !!username,
  });
}

// Private list of inbox questions (unanswered)
export function useInboxQuestions() {
  const { isAuthenticated } = useAuth();
  
  return useQuery({
    queryKey: [api.questions.listInbox.path],
    queryFn: async () => {
      const res = await fetch(api.questions.listInbox.path, { credentials: "include" });
      if (!res.ok) throw new Error("Failed to fetch inbox");
      return api.questions.listInbox.responses[200].parse(await res.json());
    },
    enabled: isAuthenticated,
  });
}

export function useCreateQuestion() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async (data: InsertQuestion) => {
      const validated = api.questions.create.input.parse(data);
      const res = await fetch(api.questions.create.path, {
        method: api.questions.create.method,
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(validated),
        credentials: "include",
      });
      
      if (!res.ok) {
        if (res.status === 400) {
          const error = await res.json();
          throw new Error(error.message || "Validation failed");
        }
        throw new Error("Failed to submit question");
      }
      
      return api.questions.create.responses[201].parse(await res.json());
    },
    onSuccess: (data) => {
      // If we are looking at our own inbox, this might show up if we ask ourselves (testing)
      queryClient.invalidateQueries({ queryKey: [api.questions.listInbox.path] });
    },
  });
}

export function useAnswerQuestion() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async ({ id, answer }: { id: number; answer: string }) => {
      const payload: AnswerQuestionRequest = { answer };
      const validated = api.questions.answer.input.parse(payload);
      const url = buildUrl(api.questions.answer.path, { id });
      
      const res = await fetch(url, {
        method: api.questions.answer.method,
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(validated),
        credentials: "include",
      });
      
      if (!res.ok) throw new Error("Failed to submit answer");
      
      return api.questions.answer.responses[200].parse(await res.json());
    },
    onSuccess: (_, variables) => {
      // Refresh inbox to remove the answered question
      queryClient.invalidateQueries({ queryKey: [api.questions.listInbox.path] });
      // We don't easily know the username here without extra data, but generally 
      // we invalidate the public feed if we could. 
      // For now, let's just invalidate the inbox.
    },
  });
}

export function useDeleteQuestion() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async (id: number) => {
      const url = buildUrl(api.questions.delete.path, { id });
      const res = await fetch(url, {
        method: api.questions.delete.method,
        credentials: "include",
      });
      
      if (!res.ok) throw new Error("Failed to delete question");
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [api.questions.listInbox.path] });
    },
  });
}

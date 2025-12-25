import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api, buildUrl } from "@shared/routes";
import type { InsertProfile } from "@shared/schema";
import { useAuth } from "./use-auth";

export function useProfile(username?: string) {
  // Use a query key that depends on the username
  // If username is undefined, this query will be disabled (enabled: !!username)
  return useQuery({
    queryKey: [api.profiles.getByUsername.path, username],
    queryFn: async () => {
      if (!username) return null;
      const url = buildUrl(api.profiles.getByUsername.path, { username });
      const res = await fetch(url, { credentials: "include" });
      
      if (res.status === 404) return null;
      if (!res.ok) throw new Error("Failed to fetch profile");
      
      return api.profiles.getByUsername.responses[200].parse(await res.json());
    },
    enabled: !!username,
  });
}

export function useMyProfile() {
  const { isAuthenticated } = useAuth();
  
  return useQuery({
    queryKey: [api.profiles.me.path],
    queryFn: async () => {
      const res = await fetch(api.profiles.me.path, { credentials: "include" });
      if (res.status === 401) return null; // Not logged in
      if (!res.ok) throw new Error("Failed to fetch my profile");
      
      const data = await res.json();
      if (!data) return null; // Logged in but no profile yet
      
      return api.profiles.me.responses[200].parse(data);
    },
    enabled: isAuthenticated,
  });
}

export function useUpdateProfile() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async (data: Omit<InsertProfile, "userId">) => {
      const validated = api.profiles.update.input.parse(data);
      const res = await fetch(api.profiles.update.path, {
        method: api.profiles.update.method,
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(validated),
        credentials: "include",
      });
      
      if (!res.ok) {
        if (res.status === 400) {
           // Try to parse Zod error from backend
           const error = await res.json();
           throw new Error(error.message || "Validation failed");
        }
        throw new Error("Failed to update profile");
      }
      
      return api.profiles.update.responses[200].parse(await res.json());
    },
    onSuccess: (updatedProfile) => {
      // Invalidate both 'me' query and the public profile query
      queryClient.invalidateQueries({ queryKey: [api.profiles.me.path] });
      queryClient.invalidateQueries({ 
        queryKey: [api.profiles.getByUsername.path, updatedProfile.username] 
      });
    },
  });
}

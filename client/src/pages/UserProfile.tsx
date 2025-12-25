import { useRoute } from "wouter";
import { useProfile } from "@/hooks/use-profiles";
import { usePublicQuestions } from "@/hooks/use-questions";
import { Navbar } from "@/components/Navbar";
import { AskForm } from "@/components/AskForm";
import { QuestionCard } from "@/components/QuestionCard";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Loader2, MessageSquareOff } from "lucide-react";

export default function UserProfile() {
  const [match, params] = useRoute("/:username");
  const username = params?.username;
  
  const { data: profile, isLoading: profileLoading, error } = useProfile(username);
  const { data: questions, isLoading: questionsLoading } = usePublicQuestions(username);

  if (profileLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    );
  }

  if (error || !profile) {
    return (
      <div className="min-h-screen flex flex-col items-center justify-center bg-background">
        <h1 className="text-2xl font-bold mb-2">User Not Found</h1>
        <p className="text-muted-foreground">The profile you are looking for does not exist.</p>
        <a href="/" className="mt-6 text-primary hover:underline">Go Home</a>
      </div>
    );
  }

  const user = profile.user;

  return (
    <div className="min-h-screen bg-background font-sans">
      <Navbar />
      
      {/* Profile Header Background */}
      <div className="h-48 md:h-64 bg-gradient-to-r from-primary/10 via-primary/5 to-secondary w-full relative overflow-hidden">
        <div className="absolute inset-0 bg-grid-white/10 [mask-image:linear-gradient(0deg,white,rgba(255,255,255,0.6))] dark:[mask-image:linear-gradient(0deg,rgba(255,255,255,0.1),rgba(255,255,255,0.5))]"></div>
      </div>

      <div className="container max-w-3xl px-4 pb-20 -mt-20 md:-mt-24 relative z-10">
        {/* Profile Card */}
        <div className="flex flex-col items-center text-center mb-10">
          <div className="relative mb-4">
            <div className="absolute -inset-1 bg-gradient-to-br from-primary to-accent rounded-full blur opacity-30"></div>
            <Avatar className="h-32 w-32 md:h-40 md:w-40 border-4 border-background shadow-xl relative">
              <AvatarImage src={user.profileImageUrl || undefined} className="object-cover" />
              <AvatarFallback className="text-4xl bg-muted text-muted-foreground font-display">
                {user.firstName?.[0]}
              </AvatarFallback>
            </Avatar>
          </div>
          
          <h1 className="text-3xl md:text-4xl font-display font-bold text-foreground mb-2">
            {user.firstName} {user.lastName}
          </h1>
          <p className="text-lg text-muted-foreground font-medium mb-4">@{profile.username}</p>
          
          {profile.bio && (
            <p className="max-w-xl text-base text-foreground/80 leading-relaxed mb-6 bg-secondary/30 px-6 py-3 rounded-2xl border border-secondary">
              {profile.bio}
            </p>
          )}
        </div>

        {/* Ask Form */}
        <div className="mb-16">
          <AskForm targetUserId={user.id} targetUsername={user.firstName || profile.username} />
        </div>

        {/* Feed Header */}
        <div className="mb-6 flex items-center justify-between border-b pb-4">
          <h2 className="text-xl font-bold font-display">Answered Questions</h2>
          <span className="text-sm font-medium text-muted-foreground bg-secondary px-3 py-1 rounded-full">
            {questions?.length || 0}
          </span>
        </div>

        {/* Questions Feed */}
        <div className="space-y-6">
          {questionsLoading ? (
            <div className="flex justify-center py-10"><Loader2 className="h-8 w-8 animate-spin text-muted-foreground" /></div>
          ) : questions && questions.length > 0 ? (
            questions.map((question) => (
              <QuestionCard key={question.id} question={question} />
            ))
          ) : (
            <div className="text-center py-16 px-4 bg-secondary/20 rounded-2xl border border-dashed border-border/60">
              <MessageSquareOff className="h-10 w-10 mx-auto text-muted-foreground/30 mb-4" />
              <p className="text-muted-foreground font-medium">No answered questions yet.</p>
              <p className="text-sm text-muted-foreground/60 mt-1">Be the first to ask something!</p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

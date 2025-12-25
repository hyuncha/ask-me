import { useEffect, useState } from "react";
import { Link, useLocation } from "wouter";
import { useAuth } from "@/hooks/use-auth";
import { useMyProfile } from "@/hooks/use-profiles";
import { useInboxQuestions, usePublicQuestions } from "@/hooks/use-questions";
import { Navbar } from "@/components/Navbar";
import { AnswerDialog } from "@/components/AnswerDialog";
import { Button } from "@/components/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Card, CardContent } from "@/components/ui/card";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Loader2, Inbox, MessageSquare, Copy, Check, ExternalLink } from "lucide-react";
import { formatDistanceToNow } from "date-fns";
import type { Question, User } from "@shared/schema";
import { useToast } from "@/hooks/use-toast";

export default function Dashboard() {
  const { isAuthenticated, isLoading: authLoading, user } = useAuth();
  const { data: profile, isLoading: profileLoading } = useMyProfile();
  const [, setLocation] = useLocation();
  const { toast } = useToast();

  const { data: inbox, isLoading: inboxLoading } = useInboxQuestions();
  // Using profile username to fetch my own public answered questions for the "Answered" tab
  const { data: answered, isLoading: answeredLoading } = usePublicQuestions(profile?.username);

  const [selectedQuestion, setSelectedQuestion] = useState<(Question & { author: User | null }) | null>(null);
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    if (!authLoading && !isAuthenticated) {
      window.location.href = "/api/login";
    }
    if (!profileLoading && !profile && isAuthenticated) {
      setLocation("/onboarding");
    }
  }, [authLoading, isAuthenticated, profileLoading, profile, setLocation]);

  if (authLoading || profileLoading || !profile) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    );
  }

  const copyLink = () => {
    const url = `${window.location.origin}/${profile.username}`;
    navigator.clipboard.writeText(url);
    setCopied(true);
    toast({
      title: "Link Copied",
      description: "Share it with your followers!",
    });
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="min-h-screen bg-secondary/20 font-sans pb-20">
      <Navbar />
      
      <div className="container max-w-4xl pt-8">
        {/* Dashboard Header */}
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-6 mb-8">
          <div>
            <h1 className="text-3xl font-display font-bold text-foreground">Dashboard</h1>
            <p className="text-muted-foreground mt-1">Manage your questions and answers.</p>
          </div>
          
          <div className="flex items-center gap-3 bg-card p-2 rounded-xl border shadow-sm">
            <Button variant="ghost" size="sm" asChild className="hover:bg-secondary">
               <Link href={`/${profile.username}`}>
                 <ExternalLink className="mr-2 h-4 w-4" />
                 View Public Page
               </Link>
            </Button>
            <div className="h-4 w-px bg-border"></div>
            <Button variant="outline" size="sm" onClick={copyLink} className="border-primary/20 text-primary hover:text-primary hover:bg-primary/5">
              {copied ? <Check className="mr-2 h-4 w-4" /> : <Copy className="mr-2 h-4 w-4" />}
              {copied ? "Copied" : "Copy Link"}
            </Button>
          </div>
        </div>

        <Tabs defaultValue="inbox" className="w-full">
          <TabsList className="grid w-full grid-cols-2 mb-8 bg-card/50 p-1 border h-12 rounded-xl">
            <TabsTrigger value="inbox" className="rounded-lg data-[state=active]:bg-primary data-[state=active]:text-primary-foreground">
              <Inbox className="mr-2 h-4 w-4" />
              Inbox
              {inbox && inbox.length > 0 && (
                <span className="ml-2 bg-primary-foreground/20 text-xs px-2 py-0.5 rounded-full">
                  {inbox.length}
                </span>
              )}
            </TabsTrigger>
            <TabsTrigger value="answered" className="rounded-lg data-[state=active]:bg-primary data-[state=active]:text-primary-foreground">
              <MessageSquare className="mr-2 h-4 w-4" />
              Answered
            </TabsTrigger>
          </TabsList>

          <TabsContent value="inbox" className="space-y-4 focus-visible:outline-none">
            {inboxLoading ? (
               <div className="py-12 flex justify-center"><Loader2 className="h-8 w-8 animate-spin text-muted-foreground" /></div>
            ) : inbox && inbox.length > 0 ? (
              inbox.map((question) => (
                <Card key={question.id} className="overflow-hidden card-hover border-l-4 border-l-primary/50">
                  <CardContent className="p-6 flex flex-col md:flex-row gap-6 md:items-center justify-between">
                    <div className="space-y-2 flex-1">
                      <div className="flex items-center gap-2 text-xs text-muted-foreground mb-1">
                         <span className="font-semibold text-foreground/80">
                           {question.author ? `${question.author.firstName}` : "Anonymous"}
                         </span>
                         <span>•</span>
                         <span>{question.createdAt && formatDistanceToNow(new Date(question.createdAt), { addSuffix: true })}</span>
                      </div>
                      <p className="text-lg font-medium leading-relaxed">{question.content}</p>
                    </div>
                    <Button onClick={() => setSelectedQuestion(question)} className="shrink-0 bg-primary hover:bg-primary/90 rounded-full px-6">
                      Answer
                    </Button>
                  </CardContent>
                </Card>
              ))
            ) : (
              <div className="text-center py-20 bg-card rounded-2xl border border-dashed border-border/60">
                <Inbox className="h-12 w-12 mx-auto text-muted-foreground/30 mb-4" />
                <h3 className="text-lg font-medium text-foreground">Inbox Empty</h3>
                <p className="text-muted-foreground">Share your link to get more questions!</p>
              </div>
            )}
          </TabsContent>

          <TabsContent value="answered" className="space-y-4 focus-visible:outline-none">
            {answeredLoading ? (
               <div className="py-12 flex justify-center"><Loader2 className="h-8 w-8 animate-spin text-muted-foreground" /></div>
            ) : answered && answered.length > 0 ? (
              answered.map((question) => (
                <Card key={question.id} className="overflow-hidden bg-card/60">
                  <CardContent className="p-6">
                    <div className="space-y-4">
                      <div className="border-l-2 border-primary/20 pl-4 py-1">
                         <p className="text-muted-foreground text-sm mb-1">Q: {question.content}</p>
                         <div className="flex items-center gap-2 text-xs text-muted-foreground/60">
                           <span>{question.author ? question.author.firstName : "Anonymous"}</span>
                         </div>
                      </div>
                      
                      <div className="bg-secondary/50 p-4 rounded-xl">
                        <p className="text-foreground">{question.answer}</p>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              ))
            ) : (
              <div className="text-center py-20 bg-card rounded-2xl border border-dashed border-border/60">
                <MessageSquare className="h-12 w-12 mx-auto text-muted-foreground/30 mb-4" />
                <h3 className="text-lg font-medium text-foreground">No Answers Yet</h3>
                <p className="text-muted-foreground">Answer questions from your inbox to see them here.</p>
              </div>
            )}
          </TabsContent>
        </Tabs>
      </div>

      {selectedQuestion && (
        <AnswerDialog 
          question={selectedQuestion} 
          isOpen={!!selectedQuestion} 
          onClose={() => setSelectedQuestion(null)} 
        />
      )}
    </div>
  );
}

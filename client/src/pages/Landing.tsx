import { Link } from "wouter";
import { useAuth } from "@/hooks/use-auth";
import { useMyProfile } from "@/hooks/use-profiles";
import { Button } from "@/components/ui/button";
import { Navbar } from "@/components/Navbar";
import { ArrowRight, MessageCircle, Sparkles, Shield, Share2 } from "lucide-react";

export default function Landing() {
  const { isAuthenticated } = useAuth();
  const { data: profile } = useMyProfile();

  return (
    <div className="min-h-screen bg-background flex flex-col font-sans">
      <Navbar />

      <main className="flex-1">
        {/* Hero Section */}
        <section className="relative overflow-hidden py-24 lg:py-32">
          {/* Background Decor */}
          <div className="absolute inset-0 z-0">
             <div className="absolute top-0 right-0 -mr-20 -mt-20 w-96 h-96 bg-primary/10 rounded-full blur-3xl opacity-50 animate-pulse"></div>
             <div className="absolute bottom-0 left-0 -ml-20 -mb-20 w-80 h-80 bg-blue-400/10 rounded-full blur-3xl opacity-50"></div>
          </div>

          <div className="container relative z-10 flex flex-col items-center text-center">
            <div className="inline-flex items-center rounded-full border border-primary/20 bg-primary/5 px-3 py-1 text-sm font-medium text-primary mb-6 animate-fade-in-up">
              <Sparkles className="mr-2 h-3.5 w-3.5" />
              The best way to answer questions
            </div>
            
            <h1 className="max-w-4xl text-5xl font-extrabold tracking-tight sm:text-6xl lg:text-7xl mb-8 bg-gradient-to-br from-foreground to-foreground/60 bg-clip-text text-transparent">
              Create your personal <br/> Q&A space
            </h1>
            
            <p className="max-w-2xl text-lg text-muted-foreground mb-10 leading-relaxed">
              AskMe gives you a dedicated page to receive questions from your audience, friends, or followers. Answer them publicly and share your knowledge.
            </p>

            <div className="flex flex-col sm:flex-row gap-4 w-full sm:w-auto">
              {isAuthenticated ? (
                profile ? (
                  <Button asChild size="lg" className="rounded-full text-lg h-14 px-8 shadow-lg shadow-primary/25">
                    <Link href="/dashboard">
                      Go to Dashboard <ArrowRight className="ml-2 h-5 w-5" />
                    </Link>
                  </Button>
                ) : (
                  <Button asChild size="lg" className="rounded-full text-lg h-14 px-8 shadow-lg shadow-primary/25">
                    <Link href="/onboarding">
                      Create Profile <ArrowRight className="ml-2 h-5 w-5" />
                    </Link>
                  </Button>
                )
              ) : (
                <Button asChild size="lg" className="rounded-full text-lg h-14 px-8 shadow-lg shadow-primary/25">
                  <a href="/api/login">
                    Get Started Free <ArrowRight className="ml-2 h-5 w-5" />
                  </a>
                </Button>
              )}
            </div>
          </div>
        </section>

        {/* Features Section */}
        <section className="py-24 bg-secondary/30 border-y border-border/50">
          <div className="container">
            <div className="grid grid-cols-1 md:grid-cols-3 gap-12">
              <div className="flex flex-col items-center text-center p-6 rounded-2xl bg-background border border-border/50 shadow-sm hover:shadow-md transition-all">
                <div className="h-14 w-14 rounded-xl bg-primary/10 flex items-center justify-center text-primary mb-6">
                  <MessageCircle className="h-7 w-7" />
                </div>
                <h3 className="text-xl font-bold mb-3">Open Questions</h3>
                <p className="text-muted-foreground">Receive questions from anyone. You choose which ones to answer publicly.</p>
              </div>

              <div className="flex flex-col items-center text-center p-6 rounded-2xl bg-background border border-border/50 shadow-sm hover:shadow-md transition-all">
                <div className="h-14 w-14 rounded-xl bg-primary/10 flex items-center justify-center text-primary mb-6">
                  <Shield className="h-7 w-7" />
                </div>
                <h3 className="text-xl font-bold mb-3">Moderation First</h3>
                <p className="text-muted-foreground">Nothing goes public until you answer it. You have full control over your feed.</p>
              </div>

              <div className="flex flex-col items-center text-center p-6 rounded-2xl bg-background border border-border/50 shadow-sm hover:shadow-md transition-all">
                <div className="h-14 w-14 rounded-xl bg-primary/10 flex items-center justify-center text-primary mb-6">
                  <Share2 className="h-7 w-7" />
                </div>
                <h3 className="text-xl font-bold mb-3">Share Anywhere</h3>
                <p className="text-muted-foreground">Share your unique profile link on Twitter, Instagram, or LinkedIn.</p>
              </div>
            </div>
          </div>
        </section>
      </main>

      <footer className="py-8 border-t bg-background">
        <div className="container flex items-center justify-between text-sm text-muted-foreground">
          <p>&copy; {new Date().getFullYear()} AskMe. All rights reserved.</p>
          <div className="flex gap-4">
             <a href="#" className="hover:text-foreground">Privacy</a>
             <a href="#" className="hover:text-foreground">Terms</a>
          </div>
        </div>
      </footer>
    </div>
  );
}

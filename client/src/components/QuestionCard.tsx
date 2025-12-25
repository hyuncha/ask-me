import { formatDistanceToNow } from "date-fns";
import type { Question, User } from "@shared/schema";
import { Card, CardContent } from "@/components/ui/card";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { MessageCircleQuestion, Quote } from "lucide-react";

interface QuestionCardProps {
  question: Question & { author: User | null };
}

export function QuestionCard({ question }: QuestionCardProps) {
  const authorName = question.author ? `${question.author.firstName} ${question.author.lastName || ''}` : "Anonymous";
  const authorInitials = question.author?.firstName?.[0] || "?";
  
  return (
    <Card className="overflow-hidden border border-border/60 bg-card hover:border-primary/30 hover:shadow-lg hover:shadow-primary/5 transition-all duration-300 group">
      <CardContent className="p-0">
        {/* Question Section */}
        <div className="p-6 bg-secondary/30">
          <div className="flex items-start gap-4">
            <Avatar className="h-10 w-10 border-2 border-background shadow-sm">
              <AvatarImage src={question.author?.profileImageUrl || undefined} />
              <AvatarFallback className="bg-muted text-muted-foreground">{authorInitials}</AvatarFallback>
            </Avatar>
            <div className="flex-1 space-y-1">
              <div className="flex items-center justify-between">
                <span className="font-semibold text-sm text-muted-foreground">
                  {authorName} asked
                </span>
                <span className="text-xs text-muted-foreground/60">
                  {question.createdAt && formatDistanceToNow(new Date(question.createdAt), { addSuffix: true })}
                </span>
              </div>
              <h3 className="font-display text-lg font-semibold leading-relaxed text-foreground">
                {question.content}
              </h3>
            </div>
          </div>
        </div>

        {/* Answer Section */}
        {question.answer && (
          <div className="p-6 pt-4 relative bg-card">
            <Quote className="absolute top-6 left-6 h-8 w-8 text-primary/10 -scale-x-100 transform" />
            <div className="pl-14">
              <p className="text-base text-foreground/90 leading-relaxed whitespace-pre-wrap">
                {question.answer}
              </p>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}

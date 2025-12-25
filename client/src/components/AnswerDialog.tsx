import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { answerQuestionSchema } from "@shared/schema";
import type { Question, User } from "@shared/schema";
import { useAnswerQuestion } from "@/hooks/use-questions";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { useToast } from "@/hooks/use-toast";
import { Loader2 } from "lucide-react";

interface AnswerDialogProps {
  question: Question & { author: User | null };
  isOpen: boolean;
  onClose: () => void;
}

type FormValues = z.infer<typeof answerQuestionSchema>;

export function AnswerDialog({ question, isOpen, onClose }: AnswerDialogProps) {
  const { toast } = useToast();
  const answerMutation = useAnswerQuestion();
  
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<FormValues>({
    resolver: zodResolver(answerQuestionSchema),
    defaultValues: {
      answer: "",
    },
  });

  const onSubmit = (data: FormValues) => {
    answerMutation.mutate(
      { id: question.id, answer: data.answer },
      {
        onSuccess: () => {
          toast({
            title: "Answered!",
            description: "Your answer has been published.",
          });
          reset();
          onClose();
        },
        onError: (err) => {
          toast({
            title: "Error",
            description: err.message,
            variant: "destructive",
          });
        },
      }
    );
  };

  const authorName = question.author ? `${question.author.firstName}` : "Anonymous";

  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="sm:max-w-[500px] p-0 overflow-hidden border-0 shadow-2xl">
        <div className="bg-secondary/40 p-6 border-b">
          <DialogHeader>
            <DialogTitle>Answer {authorName}'s Question</DialogTitle>
            <DialogDescription className="mt-2 text-base font-medium text-foreground">
              "{question.content}"
            </DialogDescription>
          </DialogHeader>
        </div>
        
        <form onSubmit={handleSubmit(onSubmit)} className="p-6 space-y-4">
          <div className="space-y-2">
            <Textarea
              placeholder="Write your answer..."
              className="min-h-[150px] resize-none text-base focus-visible:ring-primary"
              {...register("answer")}
            />
            {errors.answer && (
              <p className="text-sm text-destructive">{errors.answer.message}</p>
            )}
          </div>
          
          <DialogFooter>
            <Button variant="outline" type="button" onClick={onClose}>
              Cancel
            </Button>
            <Button 
              type="submit" 
              disabled={answerMutation.isPending}
              className="bg-primary hover:bg-primary/90"
            >
              {answerMutation.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Publish Answer
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { insertQuestionSchema } from "@shared/schema";
import { useCreateQuestion } from "@/hooks/use-questions";
import { useAuth } from "@/hooks/use-auth";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { Switch } from "@/components/ui/switch";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { SendHorizontal, Loader2, Sparkles } from "lucide-react";
import { useToast } from "@/hooks/use-toast";
import { motion, AnimatePresence } from "framer-motion";

const formSchema = z.object({
  content: z.string().min(10, "Question must be at least 10 characters").max(500, "Question is too long"),
  isAnonymous: z.boolean().default(false),
});

type FormValues = z.infer<typeof formSchema>;

interface AskFormProps {
  targetUserId: string;
  targetUsername: string;
}

export function AskForm({ targetUserId, targetUsername }: AskFormProps) {
  const { user, isAuthenticated } = useAuth();
  const { toast } = useToast();
  const createQuestion = useCreateQuestion();
  const [isFocused, setIsFocused] = useState(false);

  const {
    register,
    handleSubmit,
    reset,
    watch,
    setValue,
    formState: { errors },
  } = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      content: "",
      isAnonymous: false,
    },
  });

  const isAnonymous = watch("isAnonymous");

  const onSubmit = (data: FormValues) => {
    // If user is not logged in, they can't be authorId (handled by backend or we enforce login?)
    // Requirement said: "If logged out: Login button" in navbar, implies access is public.
    // Schema allows null authorId.
    
    // Construct payload
    const payload = {
      content: data.content,
      targetUserId,
      authorId: isAuthenticated && !data.isAnonymous ? user?.id || null : null,
    };

    createQuestion.mutate(payload, {
      onSuccess: () => {
        toast({
          title: "Sent!",
          description: "Your question has been sent successfully.",
        });
        reset();
        setIsFocused(false);
      },
      onError: (err) => {
        toast({
          title: "Error",
          description: err.message,
          variant: "destructive",
        });
      },
    });
  };

  return (
    <Card className={`border-2 transition-all duration-300 ${isFocused ? 'border-primary shadow-lg shadow-primary/10 ring-4 ring-primary/5' : 'border-border'}`}>
      <CardHeader className="pb-3">
        <CardTitle className="flex items-center gap-2 text-lg">
          <Sparkles className="w-5 h-5 text-primary" />
          Ask me anything
        </CardTitle>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <div className="relative">
            <Textarea
              placeholder={`What's on your mind? Ask ${targetUsername} a question...`}
              className="min-h-[120px] resize-none border-0 bg-secondary/50 focus-visible:ring-0 text-base placeholder:text-muted-foreground/60 rounded-xl p-4"
              {...register("content")}
              onFocus={() => setIsFocused(true)}
              onBlur={(e) => {
                if (!e.target.value) setIsFocused(false);
              }}
            />
            <AnimatePresence>
              {errors.content && (
                <motion.p
                  initial={{ opacity: 0, y: -10 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0 }}
                  className="text-sm text-destructive mt-2 absolute -bottom-6 left-1"
                >
                  {errors.content.message}
                </motion.p>
              )}
            </AnimatePresence>
          </div>

          <div className="flex items-center justify-between pt-2">
            {isAuthenticated ? (
              <div className="flex items-center gap-2 cursor-pointer" onClick={() => setValue("isAnonymous", !isAnonymous)}>
                <Switch 
                  checked={isAnonymous} 
                  onCheckedChange={(checked) => setValue("isAnonymous", checked)} 
                  id="anonymous-mode"
                />
                <Label htmlFor="anonymous-mode" className="cursor-pointer text-sm font-medium text-muted-foreground select-none">
                  Ask anonymously
                </Label>
              </div>
            ) : (
              <p className="text-xs text-muted-foreground italic">
                Log in to ask as yourself
              </p>
            )}

            <Button 
              type="submit" 
              disabled={createQuestion.isPending}
              className="rounded-full px-6 bg-primary hover:bg-primary/90 text-primary-foreground shadow-lg shadow-primary/20 hover:shadow-xl hover:shadow-primary/30 transition-all duration-300"
            >
              {createQuestion.isPending ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Sending...
                </>
              ) : (
                <>
                  Send
                  <SendHorizontal className="ml-2 h-4 w-4" />
                </>
              )}
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  );
}

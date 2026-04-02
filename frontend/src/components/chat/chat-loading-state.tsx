import { Loader2 } from "lucide-react";

interface ChatLoadingStateProps {
	message: string;
}

export function ChatLoadingState({ message }: ChatLoadingStateProps) {
	return (
		<div className="h-full w-full flex items-center justify-center">
			<div className="text-center space-y-4">
				<Loader2 className="h-4 w-4 md:h-5 md:w-5 animate-spin text-primary mx-auto" />
				<p className="text-sm md:text-base text-muted-foreground">{message}</p>
			</div>
		</div>
	);
}

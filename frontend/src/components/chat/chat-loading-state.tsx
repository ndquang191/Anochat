import { Loader2 } from "lucide-react";

interface ChatLoadingStateProps {
	message: string;
}

export function ChatLoadingState({ message }: ChatLoadingStateProps) {
	return (
		<div className="h-full w-full flex items-center justify-center">
			<div className="text-center space-y-4">
				<Loader2 className="h-12 w-12 animate-spin text-primary mx-auto" />
				<p className="text-muted-foreground">{message}</p>
			</div>
		</div>
	);
}

import {
	Tooltip,
	TooltipContent,
	TooltipTrigger,
} from "@/components/ui/tooltip";

interface ChatMessageProps {
	content: string;
	isCurrentUser: boolean;
	created_at?: number;
}

export function ChatMessage({ content, isCurrentUser, created_at }: ChatMessageProps) {
	const bubble = (
		<div
			className={`max-w-[70%] rounded-md px-3 py-2 ${
				isCurrentUser
					? "bg-primary text-primary-foreground"
					: "bg-muted text-foreground"
			}`}
		>
			<p>{content}</p>
		</div>
	);

	return (
		<div className={`flex ${isCurrentUser ? "justify-end" : "justify-start"}`}>
			{created_at ? (
				<Tooltip delayDuration={1000}>
					<TooltipTrigger asChild>{bubble}</TooltipTrigger>
					<TooltipContent side="top">
						{new Date(created_at * 1000).toLocaleTimeString([], {
							hour: "2-digit",
							minute: "2-digit",
						})}
					</TooltipContent>
				</Tooltip>
			) : (
				bubble
			)}
		</div>
	);
}

interface ChatEmptyStateProps {
	title: string;
	description: string;
}

export function ChatEmptyState({ title, description }: ChatEmptyStateProps) {
	return (
		<div className="h-full w-full flex items-center justify-center">
			<div className="text-center space-y-4">
				<p className="text-sm md:text-base text-muted-foreground">{title}</p>
				<p className="text-sm md:text-base text-muted-foreground">{description}</p>
			</div>
		</div>
	);
}

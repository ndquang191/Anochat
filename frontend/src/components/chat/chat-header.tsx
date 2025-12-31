import { Button } from "@/components/ui/button";

interface ChatHeaderProps {
	isPartnerTyping: boolean;
	onLeave: () => void;
}

export function ChatHeader({ isPartnerTyping, onLeave }: ChatHeaderProps) {
	return (
		<div className="border-b p-4 flex justify-between items-center">
			<div>
				<h3 className="font-semibold">Phòng chat</h3>
				<p className="text-sm text-muted-foreground">
					{isPartnerTyping ? "Đang nhập..." : "Đang trực tuyến"}
				</p>
			</div>
			<Button variant="outline" size="sm" onClick={onLeave}>
				Rời phòng
			</Button>
		</div>
	);
}

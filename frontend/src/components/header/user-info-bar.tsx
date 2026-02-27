import { UserDTO } from "@/types";

interface UserInfoBarProps {
	user: UserDTO;
}

export function UserInfoBar({ user }: UserInfoBarProps) {
	const { profile } = user;

	if (!profile) return null;

	return (
		<div className="absolute top-16 left-0 right-0 bg-muted/30 px-4 py-2 text-sm text-muted-foreground border-b">
			<div className="flex items-center gap-2">
				<span className="font-medium">{user.name || "Người dùng"}</span>
				{profile.age && (
					<>
						<span>•</span>
						<span>{profile.age} tuổi</span>
					</>
				)}
				{profile.is_male !== null && profile.is_male !== undefined && (
					<>
						<span>•</span>
						<span>{profile.is_male ? "Nam" : "Nữ"}</span>
					</>
				)}
			</div>
		</div>
	);
}

"use client";

import * as React from "react";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Checkbox } from "@/components/ui/checkbox";
import { toast } from "sonner";
import { ThemeToggle } from "@/components/theme-toggle";
import { MIN_AGE, MAX_AGE } from "@/types";

const CURRENT_YEAR = new Date().getFullYear();
const MIN_BIRTH_YEAR = CURRENT_YEAR - MAX_AGE;
const MAX_BIRTH_YEAR = CURRENT_YEAR - MIN_AGE;

interface UserSettingsDialogProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	initialData: {
		nickname: string;
		birthYear: number | null;
		gender: string;
	};
	onSave: (data: UserSettingsDialogProps["initialData"]) => void;
}

export function UserSettingsDialog({ open, onOpenChange, initialData, onSave }: UserSettingsDialogProps) {
	const [nickname, setNickname] = React.useState(initialData.nickname);
	const [birthYear, setBirthYear] = React.useState<number | null>(initialData.birthYear);
	const [gender, setGender] = React.useState(() => {
		const validGenders = ["male", "female"];
		return validGenders.includes(initialData.gender) ? initialData.gender : "male";
	});
	const [isLoading, setIsLoading] = React.useState(false);

	React.useEffect(() => {
		if (open) {
			setNickname(initialData.nickname);
			setBirthYear(initialData.birthYear);
			const validGenders = ["male", "female"];
			setGender(validGenders.includes(initialData.gender) ? initialData.gender : "male");
		}
	}, [open, initialData]);

	const handleSave = async () => {
		if (birthYear !== null && (birthYear < MIN_BIRTH_YEAR || birthYear > MAX_BIRTH_YEAR)) {
			toast.error(`Năm sinh phải từ ${MIN_BIRTH_YEAR} đến ${MAX_BIRTH_YEAR}`);
			return;
		}

		setIsLoading(true);
		try {
			onSave({ nickname, birthYear, gender });
			toast.success("Thông tin đã được lưu thành công!");
			onOpenChange(false);
		} catch (error) {
			const errorMessage = error instanceof Error ? error.message : "Không thể lưu thông tin. Vui lòng thử lại.";
			toast.error(errorMessage);
		} finally {
			setIsLoading(false);
		}
	};

	return (
		<Dialog open={open} onOpenChange={onOpenChange}>
			<DialogContent className="h-fit max-h-[95vh] sm:max-w-[425px] z-[9998] overflow-y-auto">
				<DialogHeader>
					<DialogTitle>Cài đặt tài khoản</DialogTitle>
					<DialogDescription>Thay đổi thông tin cá nhân của bạn tại đây.</DialogDescription>
				</DialogHeader>
				<div className="grid gap-4 py-4">
					<div className="grid grid-cols-4 items-center gap-4">
						<Label htmlFor="nickname" className="text-right">
							Nickname
						</Label>
						<Input
							id="nickname"
							value={nickname}
							onChange={(e) => setNickname(e.target.value)}
							className="col-span-3"
							placeholder="Để trống để dùng tên Google"
							maxLength={32}
						/>
					</div>
					<div className="grid grid-cols-4 items-center gap-4">
						<Label htmlFor="birthYear" className="text-right">
							Năm sinh
						</Label>
						<Input
							id="birthYear"
							type="number"
							value={birthYear ?? ""}
							onChange={(e) => setBirthYear(e.target.value ? parseInt(e.target.value) : null)}
							className="col-span-3"
							placeholder={`${MIN_BIRTH_YEAR} – ${MAX_BIRTH_YEAR}`}
							min={MIN_BIRTH_YEAR}
							max={MAX_BIRTH_YEAR}
						/>
					</div>
					<div className="grid grid-cols-4 items-center gap-4">
						<Label className="text-right">Giới tính</Label>
						<div className="col-span-3 flex gap-6">
							<div className="flex items-center space-x-2">
								<Checkbox
									id="male"
									checked={gender === "male"}
									onCheckedChange={(checked: boolean) => {
										if (checked) setGender("male");
									}}
								/>
								<Label htmlFor="male" className="flex items-center gap-1">
									Nam
								</Label>
							</div>
							<div className="flex items-center space-x-2">
								<Checkbox
									id="female"
									checked={gender === "female"}
									onCheckedChange={(checked: boolean) => {
										if (checked) setGender("female");
									}}
								/>
								<Label htmlFor="female" className="flex items-center gap-1">
									Nữ
								</Label>
							</div>
						</div>
					</div>
					<div className="grid grid-cols-4 items-center gap-4">
						<Label className="text-right">Giao diện</Label>
						<div className="col-span-3">
							<ThemeToggle />
						</div>
					</div>
				</div>
				<DialogFooter>
					<Button variant="outline" onClick={() => onOpenChange(false)} disabled={isLoading}>
						Hủy
					</Button>
					<Button onClick={handleSave} disabled={isLoading}>
						{isLoading ? "Đang lưu..." : "Lưu thay đổi"}
					</Button>
				</DialogFooter>
			</DialogContent>
		</Dialog>
	);
}

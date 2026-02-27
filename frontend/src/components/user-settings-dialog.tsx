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
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { toast } from "sonner";

interface UserSettingsDialogProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	initialData: {
		age: number | null;
		gender: string;
		city: string;
	};
	onSave: (data: UserSettingsDialogProps["initialData"]) => void;
}

export function UserSettingsDialog({ open, onOpenChange, initialData, onSave }: UserSettingsDialogProps) {
	const [age, setAge] = React.useState(initialData.age);
	const [gender, setGender] = React.useState(() => {
		const validGenders = ["male", "female"];
		return validGenders.includes(initialData.gender) ? initialData.gender : "male";
	});
	const [city, setCity] = React.useState(initialData.city);
	const [isLoading, setIsLoading] = React.useState(false);

	React.useEffect(() => {
		if (open) {
			setAge(initialData.age);
			const validGenders = ["male", "female"];
			const validGender = validGenders.includes(initialData.gender) ? initialData.gender : "male";
			setGender(validGender);
			setCity(initialData.city);
		}
	}, [open, initialData]);

	const handleSave = async () => {
		setIsLoading(true);

		try {
			onSave({
				age,
				gender,
				city,
			});

			toast.success("Thông tin đã được lưu thành công!");
			onOpenChange(false);
		} catch (error) {
			console.error("Error saving user settings:", error);
			const errorMessage =
				error instanceof Error ? error.message : "Không thể lưu thông tin. Vui lòng thử lại.";
			toast.error(errorMessage);
		} finally {
			setIsLoading(false);
		}
	};

	return (
		<Dialog open={open} onOpenChange={onOpenChange}>
			<DialogContent className="h-fit sm:max-w-[425px] z-[9998]">
				<DialogHeader>
					<DialogTitle>Cài đặt tài khoản</DialogTitle>
					<DialogDescription>Thay đổi thông tin cá nhân của bạn tại đây.</DialogDescription>
				</DialogHeader>
				<div className="grid gap-4 py-4">
					<div className="grid grid-cols-4 items-center gap-4">
						<Label htmlFor="age" className="text-right">
							Tuổi
						</Label>
						<Input
							id="age"
							type="number"
							value={age || ""}
							onChange={(e) => setAge(e.target.value ? parseInt(e.target.value) : null)}
							className="col-span-3"
							placeholder="Nhập tuổi"
							min="1"
							max="120"
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
					<div className="grid grid-cols-4 items-center gap-4 relative">
						<Label htmlFor="region" className="text-right">
							Vùng miền
						</Label>
						<Select value={city} onValueChange={(value) => setCity(value)}>
							<SelectTrigger className="col-span-3">
								<SelectValue placeholder="Chọn vùng miền" />
							</SelectTrigger>
							<SelectContent position="popper" side="top" align="start" className="z-[99999]">
								<SelectItem value="north">Miền Bắc</SelectItem>
								<SelectItem value="central">Miền Trung</SelectItem>
								<SelectItem value="south">Miền Nam</SelectItem>
							</SelectContent>
						</Select>
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

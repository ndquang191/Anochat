import Image from "next/image";
import type { HTMLAttributes } from "react";
import { cn } from "@/lib/utils";

interface BrandLogoProps extends HTMLAttributes<HTMLDivElement> {
	iconClassName?: string;
	sloganClassName?: string;
	showSlogan?: boolean;
}

export function BrandLogo({
	className,
	iconClassName,
	sloganClassName,
	showSlogan = true,
	style,
	...props
}: BrandLogoProps) {
	return (
		<div
			className={cn("flex items-center gap-3", className)}
			style={style}
			{...props}
		>
			<Image
				src="/icon.svg"
				alt="AnoChat"
				width={40}
				height={40}
				className={cn("size-10 shrink-0", iconClassName)}
			/>
			{showSlogan && (
				<span
					className={cn(
						"whitespace-nowrap text-[15px] italic leading-[1.15] tracking-wide text-muted-foreground",
						sloganClassName
					)}
					style={{
						fontFamily:
							'"Palatino Linotype", "Book Antiqua", Palatino, Georgia, serif',
					}}
				>
					<span className="block">A quiet place</span>
					<span className="block">to be heard</span>
				</span>
			)}
		</div>
	);
}

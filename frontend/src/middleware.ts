import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

export function middleware(request: NextRequest) {
    // has_session: non-HttpOnly, lives as long as refresh token (set by backend)
    // access_token: fallback for sessions created before has_session was introduced
    const token =
        request.cookies.get("has_session")?.value ||
        request.cookies.get("access_token")?.value;
    const pathname = request.nextUrl.pathname;

    const publicPaths = ["/home", "/login", "/callback", "/error", "/robots.txt", "/sitemap.xml", "/icon.svg"];
    const isPublicPath = publicPaths.some((path) => pathname === path || pathname.startsWith(`${path}/`));

    const isLoggedIn = !!token;

    if (isLoggedIn && pathname === "/login") {
        return NextResponse.redirect(new URL("/", request.url));
    }

    if (isPublicPath) {
        return NextResponse.next();
    }

    if (!isLoggedIn) {
        const loginUrl = new URL("/login", request.url);
        loginUrl.searchParams.set("redirect", pathname);
        return NextResponse.redirect(loginUrl);
    }

    return NextResponse.next();
}

export const config = {
    matcher: ["/((?!api|_next/static|_next/image|favicon.ico|icon.svg|robots.txt|sitemap.xml).*)"],
};

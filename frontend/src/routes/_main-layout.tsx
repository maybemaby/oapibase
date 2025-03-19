import { useAuth } from "@/components/auth-provider";
import { Button } from "@/components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuSeparator, DropdownMenuTrigger } from "@/components/ui/dropdown-menu";
import { useLogout } from "@/hooks/auth";
import { Link, Outlet, createFileRoute, useNavigate, useRouter } from "@tanstack/react-router";
import React, { useEffect } from "react";

export const Route = createFileRoute("/_main-layout")({
  component: MainLayout,
});

function MainLayout() {
  const {
    loggedIn
  } = useAuth();
  const logout = useLogout();


  const navigate = useNavigate();

  const router = useRouter();

  useEffect(() => {
    if (loggedIn === false)
      navigate({
        to: "/auth/login"
      })
  }, [
    loggedIn,
    router
  ])

  return (
    <div className="h-full w-full flex flex-col">
      <header className="flex p-2 pt-4 justify-between items-center gap-3 border-b-[1px]">
        <div className="flex justify-between items-center max-w-screen-xl w-full mx-auto">

          <Link to="/">
            <p className="text-2xl font-semibold mr-4">oapibase</p>
          </Link>
          <nav className="flex gap-6 items-center">
            <Link to="/profile/feeds">Edit Feeds</Link>
            <DropdownMenu>
              <DropdownMenuTrigger>
                Menu
              </DropdownMenuTrigger>
              <DropdownMenuContent>
                <DropdownMenuItem asChild>
                  <Link href="/profile">Profile</Link>
                </DropdownMenuItem>
                <DropdownMenuSeparator></DropdownMenuSeparator>
                <DropdownMenuItem onSelect={() => logout.mutate()}>
                  Logout
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </nav>
        </div>
      </header>
      <main className="flex-grow max-w-screen-xl mx-auto w-full px-2">
        <Outlet />
      </main>
    </div>
  );
}

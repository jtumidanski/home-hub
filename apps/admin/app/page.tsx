import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Separator } from "@/components/ui/separator";

export default function Home() {
  return (
    <div className="min-h-screen bg-gradient-to-br from-neutral-50 to-neutral-100 dark:from-neutral-950 dark:to-neutral-900 p-8">
      <div className="max-w-5xl mx-auto space-y-8">
        {/* Header */}
        <div className="text-center space-y-4">
          <h1 className="text-4xl font-bold tracking-tight">
            Home Hub Admin Portal
          </h1>
          <p className="text-neutral-600 dark:text-neutral-400 text-lg">
            Next.js 16 + React 19 + Tailwind CSS 4 + shadcn/ui
          </p>
        </div>

        <Separator />

        {/* Component Showcase */}
        <div className="grid gap-6 md:grid-cols-2">
          {/* Buttons Card */}
          <Card>
            <CardHeader>
              <CardTitle>Button Components</CardTitle>
              <CardDescription>
                Various button styles and variants
              </CardDescription>
            </CardHeader>
            <CardContent className="flex flex-wrap gap-2">
              <Button>Default</Button>
              <Button variant="secondary">Secondary</Button>
              <Button variant="outline">Outline</Button>
              <Button variant="ghost">Ghost</Button>
              <Button variant="destructive">Destructive</Button>
              <Button variant="link">Link</Button>
            </CardContent>
          </Card>

          {/* Form Card */}
          <Card>
            <CardHeader>
              <CardTitle>Form Components</CardTitle>
              <CardDescription>Input fields with labels</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="name">Name</Label>
                <Input id="name" placeholder="Enter your name" />
              </div>
              <div className="space-y-2">
                <Label htmlFor="email">Email</Label>
                <Input id="email" type="email" placeholder="Enter your email" />
              </div>
            </CardContent>
          </Card>

          {/* Avatar Card */}
          <Card>
            <CardHeader>
              <CardTitle>Avatar Components</CardTitle>
              <CardDescription>User avatars with fallbacks</CardDescription>
            </CardHeader>
            <CardContent className="flex gap-4">
              <Avatar>
                <AvatarImage src="https://github.com/shadcn.png" alt="User" />
                <AvatarFallback>CN</AvatarFallback>
              </Avatar>
              <Avatar>
                <AvatarFallback>JD</AvatarFallback>
              </Avatar>
              <Avatar>
                <AvatarFallback>AB</AvatarFallback>
              </Avatar>
            </CardContent>
          </Card>

          {/* Status Card */}
          <Card>
            <CardHeader>
              <CardTitle>Project Status</CardTitle>
              <CardDescription>Current implementation phase</CardDescription>
            </CardHeader>
            <CardContent className="space-y-3">
              <div className="flex justify-between items-center">
                <span className="text-sm font-medium">Phase 1: Next.js Init</span>
                <span className="text-green-600 dark:text-green-400 text-sm">✓ Complete</span>
              </div>
              <Separator />
              <div className="flex justify-between items-center">
                <span className="text-sm font-medium">Phase 2: shadcn/ui</span>
                <span className="text-green-600 dark:text-green-400 text-sm">✓ Complete</span>
              </div>
              <Separator />
              <div className="flex justify-between items-center">
                <span className="text-sm font-medium">Phase 3: Docker</span>
                <span className="text-neutral-500 text-sm">Pending</span>
              </div>
            </CardContent>
          </Card>
        </div>

        {/* Info Section */}
        <Card className="border-neutral-200 dark:border-neutral-800">
          <CardHeader>
            <CardTitle>Architecture Overview</CardTitle>
            <CardDescription>
              Home Hub microservices platform
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid gap-4 md:grid-cols-3">
              <div className="space-y-2">
                <h3 className="font-semibold text-sm">Technology Stack</h3>
                <ul className="text-sm text-neutral-600 dark:text-neutral-400 space-y-1">
                  <li>• Next.js 16</li>
                  <li>• React 19</li>
                  <li>• TypeScript 5</li>
                  <li>• Tailwind CSS 4</li>
                </ul>
              </div>
              <div className="space-y-2">
                <h3 className="font-semibold text-sm">Features</h3>
                <ul className="text-sm text-neutral-600 dark:text-neutral-400 space-y-1">
                  <li>• App Router</li>
                  <li>• Server Components</li>
                  <li>• shadcn/ui</li>
                  <li>• Docker Ready</li>
                </ul>
              </div>
              <div className="space-y-2">
                <h3 className="font-semibold text-sm">Development</h3>
                <ul className="text-sm text-neutral-600 dark:text-neutral-400 space-y-1">
                  <li>• Port: 5174</li>
                  <li>• Hot Reload</li>
                  <li>• TypeScript Strict</li>
                  <li>• ESLint</li>
                </ul>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

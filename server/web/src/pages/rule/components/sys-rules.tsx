import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs"
import { Card } from "@/components/ui/card"
import { SysRulesTable } from "./sys-rules-table"

export function SysRules() {
    return (
        <Card className="p-6">
            <Tabs defaultValue="category" className="w-full">
                <div className="bg-zinc-50 p-4">
                    <TabsList className="bg-transparent border-0 p-0 h-auto">
                        <TabsTrigger 
                            value="category"
                            className="bg-transparent px-4 py-2 text-sm font-medium border-r border-zinc-200 data-[state=active]:bg-transparent data-[state=active]:text-zinc-900 data-[state=inactive]:text-zinc-500"
                        >
                            系统规则大类
                        </TabsTrigger>
                        <TabsTrigger 
                            value="all"
                            className="bg-transparent px-4 py-2 text-sm font-medium data-[state=active]:bg-transparent data-[state=active]:text-zinc-900 data-[state=inactive]:text-zinc-500"
                        >
                            全部系统规则
                        </TabsTrigger>
                    </TabsList>
                </div>

                <TabsContent value="category" className="mt-0">
                    <SysRulesTable type="category" />
                </TabsContent>

                <TabsContent value="all" className="mt-0">
                    <SysRulesTable type="all" />
                </TabsContent>
            </Tabs>
        </Card>
    )
}
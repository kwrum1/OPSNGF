import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Checkbox } from "@/components/ui/checkbox"
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group"
import { Button } from "@/components/ui/button"
import { Settings2, MoreHorizontal } from "lucide-react"
import { cn } from "@/lib/utils"

interface SysRulesTableProps {
    type: 'category' | 'all'
}

const mockCategoryData = [
    {
        id: 1,
        status: '启用',
        ruleId: 'CE0508',
        ruleName: 'XML 实体注入-通用',
        ruleType: 'XXE'
    },
    {
        id: 2,
        status: '禁用',
        ruleId: '130508',
        ruleName: 'XML 实体注入-通用',
        ruleType: 'XXE'
    },
    {
        id: 3,
        status: '启用',
        ruleId: '150508',
        ruleName: 'XML 实体注入-通用',
        ruleType: 'XXE'
    }
]

const mockAllRulesData = [
    {
        id: 1,
        name: 'SQL 注入检测',
        status: 'disabled'
    },
    {
        id: 2,
        name: 'SQL 注入检测',
        status: 'disabled'
    },
    {
        id: 3,
        name: 'SQL 注入检测',
        status: 'observe'
    }
]

export function SysRulesTable({ type }: SysRulesTableProps) {
    if (type === 'category') {
        return (
            <div className="space-y-4">
                <div className="flex justify-between items-center bg-zinc-50 p-4">
                    <div className="flex items-center gap-4">
                        <Button variant="outline" size="sm" className="h-8">全部系统规则</Button>
                        <div className="flex items-center gap-2">
                            <input 
                                type="text" 
                                placeholder="规则 ID" 
                                className="h-8 px-3 rounded border border-zinc-200 text-sm"
                            />
                            <input 
                                type="text" 
                                placeholder="规则名称" 
                                className="h-8 px-3 rounded border border-zinc-200 text-sm"
                            />
                        </div>
                    </div>
                    <Button variant="default" size="sm" className="h-8 bg-zinc-900">
                        批量设置为
                    </Button>
                </div>

                <Table>
                    <TableHeader>
                        <TableRow className="hover:bg-transparent">
                            <TableHead className="w-12">
                                <Checkbox className="rounded-sm" />
                            </TableHead>
                            <TableHead className="text-zinc-900 font-bold">状态</TableHead>
                            <TableHead className="text-zinc-900 font-bold">规则ID</TableHead>
                            <TableHead className="text-zinc-900 font-bold">规则名称</TableHead>
                            <TableHead className="text-zinc-900 font-bold">规则大类</TableHead>
                            <TableHead className="text-zinc-900 font-bold w-8"></TableHead>
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {mockCategoryData.map((rule) => (
                            <TableRow key={rule.id} className="hover:bg-transparent">
                                <TableCell className="py-4">
                                    <Checkbox className="rounded-sm" />
                                </TableCell>
                                <TableCell className="py-4">
                                    <span className={cn(
                                        "px-2 py-1 text-xs rounded",
                                        rule.status === '启用' ? "bg-emerald-100 text-emerald-700" : "bg-red-100 text-red-700"
                                    )}>
                                        {rule.status}
                                    </span>
                                </TableCell>
                                <TableCell className="py-4">{rule.ruleId}</TableCell>
                                <TableCell className="py-4">{rule.ruleName}</TableCell>
                                <TableCell className="py-4">{rule.ruleType}</TableCell>
                                <TableCell className="py-4">
                                    <MoreHorizontal className="w-4 h-4 text-zinc-400" />
                                </TableCell>
                            </TableRow>
                        ))}
                    </TableBody>
                </Table>

                <div className="flex justify-between items-center px-4 py-2 border-t border-zinc-200">
                    <span className="text-zinc-500 text-sm">0 of 5 row(s) selected.</span>
                    <div className="flex items-center gap-4">
                        <div className="flex items-center gap-2">
                            <span className="text-zinc-900 text-sm">列每页</span>
                            <Button 
                                variant="outline" 
                                size="sm" 
                                className="h-8 border-zinc-300 text-zinc-800"
                            >
                                12
                            </Button>
                        </div>
                        <div className="flex items-center gap-2">
                            <Button 
                                variant="outline" 
                                size="sm"
                                className="h-8 border-zinc-300 text-zinc-800"
                            >
                                上一页
                            </Button>
                            <Button 
                                variant="outline" 
                                size="sm"
                                className="h-8 border-zinc-300 text-zinc-800"
                            >
                                下一页
                            </Button>
                        </div>
                        <div className="flex items-center gap-2">
                            <span className="text-zinc-900">跳至</span>
                            <span className="text-zinc-900">120 页</span>
                        </div>
                    </div>
                </div>
            </div>
        )
    }

    return (
        <div className="space-y-4">
            <div className="flex justify-between items-center bg-zinc-50 p-4 border-b border-zinc-200">
                <div className="flex items-center gap-2">
                    <Settings2 className="w-5 h-5 text-zinc-400" />
                    <span className="text-zinc-500 text-sm">批量设置为</span>
                </div>
            </div>

            <Table>
                <TableHeader>
                    <TableRow className="hover:bg-transparent">
                        <TableHead className="w-12">
                            <Checkbox className="rounded-sm" />
                        </TableHead>
                        <TableHead className="text-zinc-900 font-bold">规则名称</TableHead>
                        <TableHead className="text-zinc-900 font-bold">状态</TableHead>
                    </TableRow>
                </TableHeader>
                <TableBody>
                    {mockAllRulesData.map((rule) => (
                        <TableRow key={rule.id} className="hover:bg-transparent">
                            <TableCell className="py-4">
                                <Checkbox className="rounded-sm" />
                            </TableCell>
                            <TableCell className="py-4">{rule.name}</TableCell>
                            <TableCell className="py-4">
                                <RadioGroup 
                                    defaultValue={rule.status} 
                                    className="flex items-center gap-8"
                                >
                                    <div className="flex items-center gap-2">
                                        <RadioGroupItem 
                                            value="disabled" 
                                            id={`disabled-${rule.id}`}
                                            className={cn(
                                                "border-2 border-zinc-800",
                                                rule.status === 'disabled' && "border-zinc-800"
                                            )}
                                        />
                                        <label 
                                            htmlFor={`disabled-${rule.id}`} 
                                            className="text-zinc-600 text-sm cursor-pointer"
                                        >
                                            禁用
                                        </label>
                                    </div>
                                    <div className="flex items-center gap-2">
                                        <RadioGroupItem 
                                            value="observe" 
                                            id={`observe-${rule.id}`}
                                            className={cn(
                                                "border-2 border-zinc-800",
                                                rule.status === 'observe' && "border-zinc-800"
                                            )}
                                        />
                                        <label 
                                            htmlFor={`observe-${rule.id}`} 
                                            className="text-zinc-600 text-sm cursor-pointer"
                                        >
                                            仅观察
                                        </label>
                                    </div>
                                    <div className="flex items-center gap-2">
                                        <RadioGroupItem 
                                            value="protect" 
                                            id={`protect-${rule.id}`}
                                            className={cn(
                                                "border-2 border-zinc-800",
                                                rule.status === 'protect' && "border-zinc-800"
                                            )}
                                        />
                                        <label 
                                            htmlFor={`protect-${rule.id}`} 
                                            className="text-zinc-600 text-sm cursor-pointer"
                                        >
                                            开启防护
                                        </label>
                                    </div>
                                </RadioGroup>
                            </TableCell>
                        </TableRow>
                    ))}
                </TableBody>
            </Table>

            <div className="flex justify-between items-center px-4 py-2 border-t border-zinc-200">
                <span className="text-zinc-500 text-sm">0 of 5 row(s) selected.</span>
                <div className="flex items-center gap-4">
                    <div className="flex items-center gap-2">
                        <span className="text-zinc-900 text-sm">列每页</span>
                        <Button 
                            variant="outline" 
                            size="sm" 
                            className="h-8 border-zinc-300 text-zinc-800"
                        >
                            12
                        </Button>
                    </div>
                    <div className="flex items-center gap-2">
                        <Button 
                            variant="outline" 
                            size="sm"
                            className="h-8 border-zinc-300 text-zinc-800"
                        >
                            上一页
                        </Button>
                        <Button 
                            variant="outline" 
                            size="sm"
                            className="h-8 border-zinc-300 text-zinc-800"
                        >
                            下一页
                        </Button>
                    </div>
                    <div className="flex items-center gap-2">
                        <span className="text-zinc-900">跳至</span>
                        <span className="text-zinc-900">120 页</span>
                    </div>
                </div>
            </div>
        </div>
    )
}
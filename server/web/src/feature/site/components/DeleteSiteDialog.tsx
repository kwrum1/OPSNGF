import {
    AlertDialog,
    AlertDialogAction,
    AlertDialogCancel,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import { useDeleteSite } from "../hooks/useSites"
import { AnimatePresence, motion } from "motion/react"
import {
    dialogEnterExitAnimation,
    dialogContentAnimation,
    dialogHeaderAnimation,
    dialogContentItemAnimation
} from '@/components/ui/animation/dialog-animation'
import { useTranslation } from 'react-i18next'
import { AnimatedButton } from "@/components/ui/animation/components/animated-button"

interface DeleteSiteDialogProps {
    open: boolean
    onOpenChange: (open: boolean) => void
    siteId: string | null
}

export function DeleteSiteDialog({
    open,
    onOpenChange,
    siteId
}: DeleteSiteDialogProps) {
    const { t } = useTranslation()
    const { deleteSite, isLoading } = useDeleteSite()

    const handleDelete = () => {
        if (!siteId) return

        deleteSite(siteId, {
            onSettled: () => {
                onOpenChange(false)
            }
        })
    }

    return (
        <AlertDialog open={open} onOpenChange={onOpenChange}>
            <AnimatePresence mode="wait">
                {open && (
                    <motion.div {...dialogEnterExitAnimation}>
                        <AlertDialogContent className="p-0 overflow-hidden">
                            <motion.div {...dialogContentAnimation}>
                                <motion.div {...dialogHeaderAnimation}>
                                    <AlertDialogHeader className="p-6 pb-3">
                                        <AlertDialogTitle className="text-xl">{t('site.deleteDialog.confirmTitle')}</AlertDialogTitle>
                                        <AlertDialogDescription>
                                            {t('site.deleteDialog.confirmDescription')}
                                        </AlertDialogDescription>
                                    </AlertDialogHeader>
                                </motion.div>

                                <motion.div
                                    {...dialogContentItemAnimation}
                                    className="px-6 pb-6"
                                >
                                    <AlertDialogFooter className="mt-2 flex justify-end space-x-2">
                                        <AnimatedButton>
                                            <AlertDialogCancel>{t('site.deleteDialog.cancel')}</AlertDialogCancel>
                                        </AnimatedButton>
                                        <AnimatedButton>
                                            <AlertDialogAction
                                                onClick={handleDelete}
                                                disabled={isLoading}
                                                className="bg-red-500 hover:bg-red-600"
                                            >
                                                {isLoading ? t('site.deleteDialog.deleting') : t('site.deleteDialog.delete')}
                                            </AlertDialogAction>
                                        </AnimatedButton>
                                    </AlertDialogFooter>
                                </motion.div>
                            </motion.div>
                        </AlertDialogContent>
                    </motion.div>
                )}
            </AnimatePresence>
        </AlertDialog>
    )
} 
import { useTranslation } from 'react-i18next'

interface LoadingProps {
  message?: string
}

export function Loading({ message }: LoadingProps) {
  const { t } = useTranslation()

  return (
    <div className="flex h-full w-full items-center justify-center min-h-[50vh]">
      <div className="flex items-center space-x-2">
        <div className="h-4 w-4 animate-spin rounded-full border-2 border-gray-300 border-t-blue-600" />
        <span>{message || t('cluster.loading')}</span>
      </div>
    </div>
  )
}

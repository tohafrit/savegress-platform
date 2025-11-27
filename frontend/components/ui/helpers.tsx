'use client';

import { useState } from 'react';
import {
  HelpCircle,
  Info,
  Lightbulb,
  AlertCircle,
  CheckCircle,
  X,
  ChevronDown,
  ChevronUp,
  ExternalLink,
  BookOpen,
} from 'lucide-react';

/**
 * Tooltip - всплывающая подсказка при наведении
 */
export function Tooltip({
  children,
  content,
  position = 'top',
}: {
  children: React.ReactNode;
  content: string;
  position?: 'top' | 'bottom' | 'left' | 'right';
}) {
  const [isVisible, setIsVisible] = useState(false);

  const positionClasses = {
    top: 'bottom-full left-1/2 -translate-x-1/2 mb-2',
    bottom: 'top-full left-1/2 -translate-x-1/2 mt-2',
    left: 'right-full top-1/2 -translate-y-1/2 mr-2',
    right: 'left-full top-1/2 -translate-y-1/2 ml-2',
  };

  return (
    <div
      className="relative inline-flex"
      onMouseEnter={() => setIsVisible(true)}
      onMouseLeave={() => setIsVisible(false)}
    >
      {children}
      {isVisible && (
        <div
          className={`absolute z-50 px-3 py-2 text-sm text-white bg-[#0a1628] border border-cyan-40 rounded-lg shadow-lg whitespace-nowrap ${positionClasses[position]}`}
        >
          {content}
        </div>
      )}
    </div>
  );
}

/**
 * HelpIcon - иконка с вопросом и подсказкой
 */
export function HelpIcon({ text, size = 'sm' }: { text: string; size?: 'sm' | 'md' }) {
  return (
    <Tooltip content={text}>
      <HelpCircle
        className={`text-grey hover:text-accent-cyan cursor-help transition-colors ${
          size === 'sm' ? 'w-4 h-4' : 'w-5 h-5'
        }`}
      />
    </Tooltip>
  );
}

/**
 * InfoBanner - информационный баннер с разными типами
 */
export function InfoBanner({
  type = 'info',
  title,
  children,
  dismissible = false,
  onDismiss,
  action,
}: {
  type?: 'info' | 'tip' | 'warning' | 'success';
  title?: string;
  children: React.ReactNode;
  dismissible?: boolean;
  onDismiss?: () => void;
  action?: {
    label: string;
    href?: string;
    onClick?: () => void;
  };
}) {
  const [isDismissed, setIsDismissed] = useState(false);

  if (isDismissed) return null;

  const styles = {
    info: {
      bg: 'bg-accent-blue/10',
      border: 'border-accent-blue/30',
      icon: Info,
      iconColor: 'text-accent-blue',
      titleColor: 'text-accent-blue',
    },
    tip: {
      bg: 'bg-accent-cyan/10',
      border: 'border-accent-cyan/30',
      icon: Lightbulb,
      iconColor: 'text-accent-cyan',
      titleColor: 'text-accent-cyan',
    },
    warning: {
      bg: 'bg-accent-orange/10',
      border: 'border-accent-orange/30',
      icon: AlertCircle,
      iconColor: 'text-accent-orange',
      titleColor: 'text-accent-orange',
    },
    success: {
      bg: 'bg-green-500/10',
      border: 'border-green-500/30',
      icon: CheckCircle,
      iconColor: 'text-green-400',
      titleColor: 'text-green-400',
    },
  };

  const style = styles[type];
  const Icon = style.icon;

  const handleDismiss = () => {
    setIsDismissed(true);
    onDismiss?.();
  };

  return (
    <div className={`${style.bg} ${style.border} border rounded-xl p-4`}>
      <div className="flex items-start gap-3">
        <Icon className={`w-5 h-5 ${style.iconColor} flex-shrink-0 mt-0.5`} />
        <div className="flex-1 min-w-0">
          {title && (
            <h4 className={`font-medium ${style.titleColor} mb-1`}>{title}</h4>
          )}
          <div className="text-sm text-grey">{children}</div>
          {action && (
            <div className="mt-3">
              {action.href ? (
                <a
                  href={action.href}
                  className={`inline-flex items-center gap-1 text-sm font-medium ${style.iconColor} hover:underline`}
                >
                  {action.label}
                  <ExternalLink className="w-3 h-3" />
                </a>
              ) : (
                <button
                  onClick={action.onClick}
                  className={`text-sm font-medium ${style.iconColor} hover:underline`}
                >
                  {action.label}
                </button>
              )}
            </div>
          )}
        </div>
        {dismissible && (
          <button
            onClick={handleDismiss}
            className="text-grey hover:text-white transition-colors"
          >
            <X className="w-4 h-4" />
          </button>
        )}
      </div>
    </div>
  );
}

/**
 * FeatureHighlight - выделение фичи с описанием
 */
export function FeatureHighlight({
  icon: Icon,
  title,
  description,
  badge,
}: {
  icon: React.ElementType;
  title: string;
  description: string;
  badge?: string;
}) {
  return (
    <div className="flex items-start gap-3 p-4 rounded-lg bg-primary-dark/50 border border-cyan-40/30">
      <div className="p-2 rounded-lg bg-accent-cyan/10">
        <Icon className="w-5 h-5 text-accent-cyan" />
      </div>
      <div className="flex-1">
        <div className="flex items-center gap-2">
          <h4 className="font-medium text-white">{title}</h4>
          {badge && (
            <span className="px-2 py-0.5 text-xs font-medium rounded-full bg-accent-cyan/20 text-accent-cyan">
              {badge}
            </span>
          )}
        </div>
        <p className="text-sm text-grey mt-1">{description}</p>
      </div>
    </div>
  );
}

/**
 * StepIndicator - индикатор шагов
 */
export function StepIndicator({
  steps,
  currentStep,
}: {
  steps: { title: string; description?: string }[];
  currentStep: number;
}) {
  return (
    <div className="space-y-4">
      {steps.map((step, index) => (
        <div key={index} className="flex items-start gap-4">
          <div
            className={`w-8 h-8 rounded-full flex items-center justify-center text-sm font-medium flex-shrink-0 ${
              index < currentStep
                ? 'bg-accent-cyan text-white'
                : index === currentStep
                ? 'bg-gradient-btn-primary text-white'
                : 'bg-primary-dark text-grey border border-cyan-40'
            }`}
          >
            {index < currentStep ? '✓' : index + 1}
          </div>
          <div className="flex-1 pt-1">
            <h4
              className={`font-medium ${
                index <= currentStep ? 'text-white' : 'text-grey'
              }`}
            >
              {step.title}
            </h4>
            {step.description && (
              <p className="text-sm text-grey mt-0.5">{step.description}</p>
            )}
          </div>
        </div>
      ))}
    </div>
  );
}

/**
 * ExpandableSection - раскрывающаяся секция
 */
export function ExpandableSection({
  title,
  children,
  defaultExpanded = false,
  icon: Icon,
}: {
  title: string;
  children: React.ReactNode;
  defaultExpanded?: boolean;
  icon?: React.ElementType;
}) {
  const [isExpanded, setIsExpanded] = useState(defaultExpanded);

  return (
    <div className="border border-cyan-40/30 rounded-xl overflow-hidden">
      <button
        onClick={() => setIsExpanded(!isExpanded)}
        className="w-full px-4 py-3 flex items-center justify-between bg-primary-dark/30 hover:bg-primary-dark/50 transition-colors"
      >
        <div className="flex items-center gap-3">
          {Icon && <Icon className="w-5 h-5 text-accent-cyan" />}
          <span className="font-medium text-white">{title}</span>
        </div>
        {isExpanded ? (
          <ChevronUp className="w-5 h-5 text-grey" />
        ) : (
          <ChevronDown className="w-5 h-5 text-grey" />
        )}
      </button>
      {isExpanded && (
        <div className="px-4 py-4 border-t border-cyan-40/30">{children}</div>
      )}
    </div>
  );
}

/**
 * QuickGuide - быстрый гайд
 */
export function QuickGuide({
  title,
  steps,
  learnMoreHref,
}: {
  title: string;
  steps: string[];
  learnMoreHref?: string;
}) {
  return (
    <div className="bg-gradient-to-br from-accent-cyan/5 to-accent-blue/5 border border-cyan-40/30 rounded-xl p-5">
      <div className="flex items-center gap-2 mb-4">
        <BookOpen className="w-5 h-5 text-accent-cyan" />
        <h3 className="font-medium text-white">{title}</h3>
      </div>
      <ol className="space-y-3">
        {steps.map((step, index) => (
          <li key={index} className="flex items-start gap-3">
            <span className="w-6 h-6 rounded-full bg-accent-cyan/20 text-accent-cyan text-sm font-medium flex items-center justify-center flex-shrink-0">
              {index + 1}
            </span>
            <span className="text-sm text-grey pt-0.5">{step}</span>
          </li>
        ))}
      </ol>
      {learnMoreHref && (
        <a
          href={learnMoreHref}
          className="inline-flex items-center gap-1 text-sm text-accent-cyan hover:text-accent-cyan-bright mt-4 transition-colors"
        >
          Learn more
          <ExternalLink className="w-3 h-3" />
        </a>
      )}
    </div>
  );
}

/**
 * EmptyStateWithGuide - пустое состояние с гайдом
 */
export function EmptyStateWithGuide({
  icon: Icon,
  title,
  description,
  guide,
  action,
}: {
  icon: React.ElementType;
  title: string;
  description: string;
  guide?: {
    title: string;
    steps: string[];
  };
  action?: {
    label: string;
    onClick: () => void;
  };
}) {
  return (
    <div className="card-dark p-8">
      <div className="max-w-2xl mx-auto">
        <div className="text-center mb-8">
          <div className="w-16 h-16 rounded-full bg-primary-dark flex items-center justify-center mx-auto mb-4">
            <Icon className="w-8 h-8 text-grey" />
          </div>
          <h3 className="text-xl font-medium text-white mb-2">{title}</h3>
          <p className="text-grey">{description}</p>
        </div>

        {guide && (
          <div className="mb-8">
            <QuickGuide title={guide.title} steps={guide.steps} />
          </div>
        )}

        {action && (
          <div className="text-center">
            <button onClick={action.onClick} className="btn-primary px-6 py-3">
              {action.label}
            </button>
          </div>
        )}
      </div>
    </div>
  );
}

/**
 * FormFieldWithHelp - поле формы с подсказкой
 */
export function FormFieldWithHelp({
  label,
  help,
  tip,
  required,
  children,
}: {
  label: string;
  help?: string;
  tip?: string;
  required?: boolean;
  children: React.ReactNode;
}) {
  return (
    <div>
      <div className="flex items-center gap-2 mb-2">
        <label className="text-sm font-medium text-grey">
          {label}
          {required && <span className="text-accent-orange ml-1">*</span>}
        </label>
        {help && <HelpIcon text={help} />}
      </div>
      {children}
      {tip && (
        <p className="text-xs text-grey mt-1.5 flex items-start gap-1.5">
          <Lightbulb className="w-3 h-3 mt-0.5 text-accent-cyan flex-shrink-0" />
          {tip}
        </p>
      )}
    </div>
  );
}

/**
 * PageHeader - заголовок страницы с описанием
 */
export function PageHeader({
  title,
  description,
  tip,
  action,
}: {
  title: string;
  description: string;
  tip?: string;
  action?: React.ReactNode;
}) {
  return (
    <div className="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-4">
      <div className="space-y-1">
        <h1 className="text-h4 text-white">{title}</h1>
        <p className="text-content-1 text-grey">{description}</p>
        {tip && (
          <p className="text-sm text-accent-cyan flex items-center gap-1.5 mt-2">
            <Lightbulb className="w-4 h-4" />
            {tip}
          </p>
        )}
      </div>
      {action && <div className="flex-shrink-0">{action}</div>}
    </div>
  );
}

/**
 * MetricCard - карточка метрики с объяснением
 */
export function MetricCard({
  title,
  value,
  subtitle,
  icon: Icon,
  help,
  trend,
  color = 'text-accent-cyan',
}: {
  title: string;
  value: string | number;
  subtitle: string;
  icon: React.ElementType;
  help?: string;
  trend?: { value: number; isPositive: boolean };
  color?: string;
}) {
  return (
    <div className="card-dark p-5">
      <div className="flex items-center justify-between mb-3">
        <div className="flex items-center gap-2">
          <span className="text-sm text-grey">{title}</span>
          {help && <HelpIcon text={help} />}
        </div>
        <div className="p-2 rounded-lg bg-primary-dark">
          <Icon className={`w-5 h-5 ${color}`} />
        </div>
      </div>
      <div className="flex items-end justify-between">
        <div>
          <p className="text-2xl font-bold text-white">{value}</p>
          <p className="text-sm text-grey mt-1">{subtitle}</p>
        </div>
        {trend && (
          <div
            className={`text-sm font-medium ${
              trend.isPositive ? 'text-green-400' : 'text-red-400'
            }`}
          >
            {trend.isPositive ? '↑' : '↓'} {Math.abs(trend.value)}%
          </div>
        )}
      </div>
    </div>
  );
}

/**
 * WelcomeBanner - приветственный баннер для новых пользователей
 */
export function WelcomeBanner({
  userName,
  onDismiss,
  onGetStarted,
}: {
  userName?: string;
  onDismiss?: () => void;
  onGetStarted?: () => void;
}) {
  const [isDismissed, setIsDismissed] = useState(false);

  if (isDismissed) return null;

  return (
    <div className="bg-gradient-to-r from-accent-cyan/10 via-accent-blue/10 to-accent-cyan/10 border border-accent-cyan/30 rounded-xl p-6">
      <div className="flex items-start justify-between">
        <div className="flex-1">
          <h2 className="text-xl font-semibold text-white mb-2">
            {userName ? `Welcome back, ${userName}! 👋` : 'Welcome to Savegress! 👋'}
          </h2>
          <p className="text-grey mb-4 max-w-2xl">
            Savegress helps you replicate data from your databases in real-time.
            Set up a connection, create a pipeline, and start streaming changes to any destination.
          </p>
          <div className="flex items-center gap-4">
            {onGetStarted && (
              <button onClick={onGetStarted} className="btn-primary px-5 py-2.5 text-sm">
                Get Started
              </button>
            )}
            <a href="/docs" className="text-sm text-accent-cyan hover:text-accent-cyan-bright transition-colors">
              Read the docs →
            </a>
          </div>
        </div>
        {onDismiss && (
          <button
            onClick={() => {
              setIsDismissed(true);
              onDismiss();
            }}
            className="text-grey hover:text-white transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        )}
      </div>
    </div>
  );
}

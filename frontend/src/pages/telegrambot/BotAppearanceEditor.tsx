import { useState } from 'react';
import { Button, Card, Input, Segmented, Select, Space, Switch } from 'antd';
import { useTranslation } from 'react-i18next';
import { CheckOutlined, RobotOutlined } from '@ant-design/icons';
import type { BotPresentation } from '@/generated/types';
import { PANEL_ACCENTS, useTheme } from '@/hooks/useTheme';

export const emptyPresentation: BotPresentation = {
  brandName: '',
  welcomeEN: '',
  announcement: '',
  supportURL: '',
  tutorialURL: '',
  defaultLanguage: 'auto',
  menuLayout: 'spacious',
  hideTrial: false,
  hideSupport: false,
  hideTutorial: false,
  labels: {},
};

const actions = ['buy', 'trial', 'invoices', 'subs', 'account', 'support', 'tutorial'] as const;
const fallbackLabels = {
  en: {
    buy: '🛍 Buy new subscription',
    trial: '🆓 Get free trial',
    invoices: '📄 My invoices',
    subs: '💎 My subscriptions',
    account: '👤 My account',
    support: '📞 Support',
    tutorial: '📱 Connection guide',
  },
  fa: {
    buy: '🛍 خرید اشتراک جدید',
    trial: '🆓 دریافت اشتراک تست',
    invoices: '📄 فاکتورهای من',
    subs: '💎 اشتراک‌های من',
    account: '👤 حساب کاربری',
    support: '📞 پشتیبانی',
    tutorial: '📱 آموزش اتصال',
  },
};

interface Props {
  value: BotPresentation;
  onChange: (value: BotPresentation) => void;
  welcome: string;
  onWelcomeChange: (text: string) => void;
  username: string;
  saving: boolean;
  onSave: () => void;
}

export default function BotAppearanceEditor({
  value,
  onChange,
  welcome,
  onWelcomeChange,
  username,
  saving,
  onSave,
}: Props) {
  const { t } = useTranslation();
  const { accent, setAccent, compact, setCompact } = useTheme();
  const [language, setLanguage] = useState<'fa' | 'en'>('fa');
  const p = { ...emptyPresentation, ...value, labels: value.labels || {} };
  const change = <K extends keyof BotPresentation>(key: K, next: BotPresentation[K]) =>
    onChange({ ...p, [key]: next });
  const label = (action: (typeof actions)[number]) =>
    p.labels[`${language}:${action}`]?.trim() || fallbackLabels[language][action];
  const hidden = (action: string) =>
    (action === 'trial' && p.hideTrial) ||
    (action === 'support' && p.hideSupport) ||
    (action === 'tutorial' && p.hideTutorial);
  const rows =
    p.menuLayout === 'compact'
      ? [['buy', 'subs'], ['trial', 'invoices'], ['account'], ['support', 'tutorial']]
      : [['buy'], ['trial'], ['invoices', 'subs'], ['account'], ['support', 'tutorial']];
  const textField = (
    key: 'brandName' | 'announcement' | 'supportURL' | 'tutorialURL',
    limit: number,
  ) => (
    <label className="bot-field" htmlFor={`bot-${key}`}>
      <span>{t(`pages.telegramBot.design.${key}`)}</span>
      <Input
        id={`bot-${key}`}
        value={p[key]}
        maxLength={limit}
        onChange={(e) => change(key, e.target.value)}
        dir="auto"
        type={key.endsWith('URL') ? 'url' : 'text'}
      />
    </label>
  );
  return (
    <div className="bot-design-grid">
      <div className="bot-design-controls">
        <Card title={t('pages.telegramBot.design.identity')}>
          <div className="bot-field-grid">
            {textField('brandName', 80)}
            <label className="bot-field">
              <span>{t('pages.telegramBot.design.language')}</span>
              <Select
                value={p.defaultLanguage || 'auto'}
                onChange={(v) => change('defaultLanguage', v)}
                options={[
                  { value: 'auto', label: t('pages.telegramBot.design.askLanguage') },
                  { value: 'fa', label: t('pages.telegramBot.design.persian') },
                  { value: 'en', label: t('pages.telegramBot.design.english') },
                ]}
              />
            </label>
          </div>
          <label className="bot-field" htmlFor="bot-welcome-fa">
            <span>{t('pages.telegramBot.design.welcomeFA')}</span>
            <Input.TextArea
              id="bot-welcome-fa"
              dir="rtl"
              rows={3}
              maxLength={2000}
              showCount
              value={welcome}
              onChange={(e) => onWelcomeChange(e.target.value)}
            />
          </label>
          <label className="bot-field" htmlFor="bot-welcome-en">
            <span>{t('pages.telegramBot.design.welcomeEN')}</span>
            <Input.TextArea
              id="bot-welcome-en"
              dir="ltr"
              rows={3}
              maxLength={2000}
              showCount
              value={p.welcomeEN}
              onChange={(e) => change('welcomeEN', e.target.value)}
            />
          </label>
          {textField('announcement', 600)}
          <p className="bot-field-help">{t('pages.telegramBot.design.plainText')}</p>
        </Card>
        <Card title={t('pages.telegramBot.design.menu')}>
          <label className="bot-field">
            <span>{t('pages.telegramBot.design.layout')}</span>
            <Segmented
              block
              value={p.menuLayout || 'spacious'}
              onChange={(v) => change('menuLayout', String(v))}
              options={[
                { value: 'spacious', label: t('pages.telegramBot.design.spacious') },
                { value: 'compact', label: t('pages.telegramBot.design.compact') },
              ]}
            />
          </label>
          <div className="bot-menu-toggles">
            {(['trial', 'support', 'tutorial'] as const).map((key) => {
              const field = {
                trial: 'hideTrial',
                support: 'hideSupport',
                tutorial: 'hideTutorial',
              }[key] as 'hideTrial' | 'hideSupport' | 'hideTutorial';
              return (
                <label key={key}>
                  <span>{t(`pages.telegramBot.design.${key}`)}</span>
                  <Switch checked={!p[field]} onChange={(v) => change(field, !v)} />
                </label>
              );
            })}
          </div>
          <div className="bot-label-heading">
            <span>{t('pages.telegramBot.design.buttonLabels')}</span>
            <Segmented
              value={language}
              onChange={setLanguage}
              options={[
                { value: 'fa', label: t('pages.telegramBot.design.persian') },
                { value: 'en', label: t('pages.telegramBot.design.english') },
              ]}
            />
          </div>
          <div className="bot-field-grid">
            {actions.map((action) => (
              <label className="bot-field" key={action} htmlFor={`bot-label-${action}`}>
                <span>{t(`pages.telegramBot.design.${action}`)}</span>
                <Input
                  id={`bot-label-${action}`}
                  dir={language === 'fa' ? 'rtl' : 'ltr'}
                  value={p.labels[`${language}:${action}`] || ''}
                  placeholder={fallbackLabels[language][action]}
                  maxLength={40}
                  onChange={(e) =>
                    change('labels', { ...p.labels, [`${language}:${action}`]: e.target.value })
                  }
                />
              </label>
            ))}
          </div>
          <Button onClick={() => change('labels', {})}>
            {t('pages.telegramBot.design.resetLabels')}
          </Button>
        </Card>
        <Card title={t('pages.telegramBot.design.links')}>
          <div className="bot-field-grid">
            {textField('supportURL', 512)}
            {textField('tutorialURL', 512)}
          </div>
          <p className="bot-field-help">{t('pages.telegramBot.design.linksHelp')}</p>
        </Card>
        <Card title={t('pages.telegramBot.design.panelStyle')}>
          <Space wrap className="bot-accent-picker">
            {Object.entries(PANEL_ACCENTS).map(([key, color]) => (
              <button
                key={key}
                type="button"
                aria-label={t(`pages.telegramBot.design.accent.${key}`)}
                aria-pressed={accent === key}
                style={{ background: color }}
                onClick={() => setAccent(key as keyof typeof PANEL_ACCENTS)}
              >
                {accent === key && <CheckOutlined />}
              </button>
            ))}
          </Space>
          <label className="bot-density">
            <span>{t('pages.telegramBot.design.density')}</span>
            <Switch checked={compact} onChange={setCompact} />
          </label>
          <p className="bot-field-help">{t('pages.telegramBot.design.localStyle')}</p>
        </Card>
        <div className="bot-save-bar">
          <Button type="primary" loading={saving} onClick={onSave}>
            {t('pages.telegramBot.design.save')}
          </Button>
        </div>
      </div>
      <aside className="bot-preview-column">
        <div className="bot-preview-caption">
          <span>{t('pages.telegramBot.design.preview')}</span>
          <span>
            {t(
              language === 'fa'
                ? 'pages.telegramBot.design.persian'
                : 'pages.telegramBot.design.english',
            )}
          </span>
        </div>
        <div className="bot-phone" dir={language === 'fa' ? 'rtl' : 'ltr'}>
          <div className="bot-phone-header">
            <span className="bot-avatar">
              <RobotOutlined />
            </span>
            <div>
              <strong>{p.brandName || username || t('pages.telegramBot.design.shop')}</strong>
              <small>{username ? `@${username}` : t('menu.telegramBot')}</small>
            </div>
          </div>
          <div className="bot-chat">
            <div className="bot-chat-bubble">
              {p.brandName && <strong>{p.brandName}</strong>}
              <p>
                {(language === 'en' ? p.welcomeEN : welcome) ||
                  (language === 'fa'
                    ? 'لطفاً یکی از گزینه‌های زیر را انتخاب کنید:'
                    : 'Please choose one of the options below:')}
              </p>
              {p.announcement && <p className="bot-chat-announcement">{p.announcement}</p>}
              <span className="bot-message-time">12:00 ✓✓</span>
            </div>
          </div>
          <div className="bot-phone-menu">
            {rows.map((row, i) => {
              const visible = row.filter((a) => !hidden(a));
              return visible.length ? (
                <div className="bot-phone-menu-row" key={i}>
                  {visible.map((action) => (
                    <span key={action}>{label(action as (typeof actions)[number])}</span>
                  ))}
                </div>
              ) : null;
            })}
          </div>
        </div>
        <p className="bot-field-help">{t('pages.telegramBot.design.previewHelp')}</p>
      </aside>
    </div>
  );
}

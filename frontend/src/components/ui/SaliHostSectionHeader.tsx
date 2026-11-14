import { useTranslation } from 'react-i18next';

type Section = 'clients' | 'inbounds' | 'nodes' | 'hosts' | 'groups' | 'settings' | 'reseller';

interface SectionDetails {
  number: string;
  category: string;
  fa: string;
  en: string;
  descriptionFa: string;
  descriptionEn: string;
}

const SECTIONS: Record<Section, SectionDetails> = {
  clients: {
    number: '01',
    category: 'CUSTOMER HUB',
    fa: 'مرکز مدیریت کاربران',
    en: 'Customer Hub',
    descriptionFa: 'حساب‌ها، اشتراک‌ها و مصرف کاربران را در یک نما مدیریت کنید.',
    descriptionEn: 'Manage accounts, subscriptions and usage in one workspace.',
  },
  inbounds: {
    number: '02',
    category: 'CONNECTION SERVICES',
    fa: 'سرویس‌های ارتباطی',
    en: 'Connection Services',
    descriptionFa: 'ورودی‌ها، پروتکل‌ها و جریان سرویس‌ها را کنترل کنید.',
    descriptionEn: 'Control inbound services, protocols and connection traffic.',
  },
  nodes: {
    number: '03',
    category: 'CLOUD NODES',
    fa: 'زیرساخت سرورها',
    en: 'Cloud Nodes',
    descriptionFa: 'نودها و وضعیت اتصال سرورهای شبکه را زیر نظر بگیرید.',
    descriptionEn: 'Monitor nodes and connected server availability.',
  },
  hosts: {
    number: '04',
    category: 'HOST MANAGEMENT',
    fa: 'مدیریت دامنه‌ها و هاست‌ها',
    en: 'Host Management',
    descriptionFa: 'هاست‌ها و آدرس‌های سرویس را از این بخش مدیریت کنید.',
    descriptionEn: 'Manage service hosts and connection addresses.',
  },
  groups: {
    number: '05',
    category: 'ACCESS GROUPS',
    fa: 'گروه‌بندی کاربران',
    en: 'Access Groups',
    descriptionFa: 'کاربران را سازمان‌دهی و دسترسی‌ها را مدیریت کنید.',
    descriptionEn: 'Organize customers and manage their access groups.',
  },
  settings: {
    number: '06',
    category: 'PLATFORM SETTINGS',
    fa: 'تنظیمات مرکز کنترل',
    en: 'Control Center Settings',
    descriptionFa: 'امنیت، اعلان‌ها و رفتار سامانه را شخصی‌سازی کنید.',
    descriptionEn: 'Configure security, notifications and platform behavior.',
  },
  reseller: {
    number: '07',
    category: 'SALES WORKSPACE',
    fa: 'داشبورد فروش و نمایندگی',
    en: 'Sales Workspace',
    descriptionFa: 'وضعیت مشتریان، اشتراک‌ها و فروش را در یک نمای یکپارچه ببینید.',
    descriptionEn: 'Review customer activity, subscriptions and sales.',
  },
};

export default function SaliHostSectionHeader({ section }: { section: Section }) {
  const { i18n } = useTranslation();
  const persian = i18n.language.startsWith('fa');
  const current = SECTIONS[section];

  return (
    <header className="salihost-section-header" data-section={section}>
      <div className="salihost-section-number" aria-hidden="true">{current.number}</div>
      <div className="salihost-section-details">
        <div className="salihost-section-eyebrow">SALIHOST.IR / {current.category}</div>
        <h1>{persian ? current.fa : current.en}</h1>
        <p>{persian ? current.descriptionFa : current.descriptionEn}</p>
      </div>
      <div className="salihost-section-wordmark" aria-hidden="true">
        <span>SH</span>
        <small>CONTROL CLOUD</small>
      </div>
    </header>
  );
}

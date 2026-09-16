import i18next from 'i18next'
import { initReactI18next } from 'react-i18next'
import LanguageDetector from 'i18next-browser-languagedetector'

import enCommon from '@/locales/en/common.json'
import enPrinters from '@/locales/en/printers.json'
import enOdoo from '@/locales/en/odoo.json'
import enSystem from '@/locales/en/system.json'

import esCommon from '@/locales/es/common.json'
import esPrinters from '@/locales/es/printers.json'
import esOdoo from '@/locales/es/odoo.json'
import esSystem from '@/locales/es/system.json'

export const LANGUAGES = [
  { code: 'es', short: 'ES', label: 'Español' },
  { code: 'en', short: 'EN', label: 'English' },
] as const

i18next
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    resources: {
      en: { common: enCommon, printers: enPrinters, odoo: enOdoo, system: enSystem },
      es: { common: esCommon, printers: esPrinters, odoo: esOdoo, system: esSystem },
    },
    fallbackLng: 'es',
    supportedLngs: LANGUAGES.map((language) => language.code),
    defaultNS: 'common',
    interpolation: { escapeValue: false },
    detection: {
      order: ['localStorage', 'navigator'],
      caches: ['localStorage'],
      lookupLocalStorage: 'dooprint.language',
    },
  })

export default i18next

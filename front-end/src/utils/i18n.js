import i18n from 'i18next'
import LanguageDetector from 'i18next-browser-languagedetector'
import { initReactI18next } from 'react-i18next'
import translations from './translations'
i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    fallbackLng: ['en', 'dev'],
    resources: {
      en: translations.en,
      fr: translations.fr,
      de: translations.de,
      it: translations.it,
      fa: translations.fa,
      hi: translations.hi,
      es: translations.es
    },
    defaultNS: 'common',
    debug: false,
    interpolation: {
      escapeValue: false
    },
    react: {
      wait: false,
      bindI18n: 'languageChanged loaded',
      bindStore: 'added removed',
      nsMode: 'default'
    }
  })
export default i18n

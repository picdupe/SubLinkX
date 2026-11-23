// translate router.meta.title, be used in breadcrumb sidebar tagsview
import i18n from "@/lang/index";

const routeTitleAliases: Record<string, string> = {
  "\u4e2d\u8f6c\u7ba1\u7406": "relayManagement",
};

export function translateRouteTitle(title: any) {
  // 判断是否存在国际化配置，如果没有原生返回
  const routeKey = routeTitleAliases[title] || title;
  const hasKey = i18n.global.te("route." + routeKey);
  if (hasKey) {
    const translatedTitle = i18n.global.t("route." + routeKey);
    return translatedTitle;
  }
  return title;
}

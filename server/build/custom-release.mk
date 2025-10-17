ifneq ($(origin CUSTOMIZE_SOURCE_DIR), undefined)
$(error CUSTOMIZE_SOURCE_DIR is already set (origin=$(origin CUSTOMIZE_SOURCE_DIR)))
endif

CUSTOMIZE_SOURCE_DIR = '$(BUILD_WEBAPP_DIR)/channels/dist'

customize-assets:
	echo "[DEBUG] Customizing web app assets for custom release..."
	echo "DIST_PATH = $(DIST_PATH)"
	echo "BUILD_WEBAPP_DIR = $(BUILD_WEBAPP_DIR)"
	echo "CUSTOM_SERVICE_NAME = $(CUSTOM_SERVICE_NAME)"
	echo "CUSTOM_PLATFORM_NAME = $(CUSTOM_PLATFORM_NAME)"
	echo "CUSTOM_JP_PLATFORM_NAME = $(CUSTOM_JP_PLATFORM_NAME)"
	echo "CUSTOMIZE_SOURCE_DIR = $(CUSTOMIZE_SOURCE_DIR)"
	pwd
	ls -l

	# Replace strings using variables
	sed -i '' -e '/"about\.notice"/!{ /"about\.copyright"/!s/Mattermost/$(CUSTOM_JP_PLATFORM_NAME)/g; }' $(CUSTOMIZE_SOURCE_DIR)/i18n/ja.*.json || true
	sed -i '' -e 's/GitLab/$(CUSTOM_SERVICE_NAME)/g' -e 's/{service}/$(CUSTOM_SERVICE_NAME)/g' -e '/"about\.notice"/!{ /"about\.copyright"/!s/Mattermost/$(CUSTOM_PLATFORM_NAME)/g; }' $(CUSTOMIZE_SOURCE_DIR)/i18n/*.json || true
	sed -i '' -e 's/Mattermost/$(CUSTOM_JP_PLATFORM_NAME)/g' $(CUSTOMIZE_SOURCE_DIR)/i18n/ja.json || true
	sed -i '' -e 's/{{.Service}}/$(CUSTOM_SERVICE_NAME)/g' -e 's/Mattermost/$(CUSTOM_PLATFORM_NAME)/g' $(CUSTOMIZE_SOURCE_DIR)/i18n/*.json || true

	# Remove GitLab icon from login screen
	icon_str='"svg",{width:"[0-9]\+",height:"[0-9]\+",viewBox:"0 0 [0-9]\+ [0-9]\+",fill:"none",xmlns:"http:\/\/www.w3.org\/2000\/svg","aria-label":t({id:"generic_icons.login.gitlab",defaultMessage:"Gitlab Icon"})}'
	file=$$(grep -l $${icon_str} $(CUSTOMIZE_SOURCE_DIR)/*.js || true)
	if [ -n "$$file" ]; then \
		sed -i '' "s|$${icon_str}|\"span\",{}|g" "$$file"; \
		sed -i '' "s/external-login-button-label//g" "$$file"; \
	fi

	# Hide Mattermost logo at the top left (before login)
	hfroute_header='o().createElement("div",{className:c()("hfroute-header",{"has-free-banner":r,"has-custom-site-name":b})}'
	file_hfroute_header=$$(grep -l "$${hfroute_header}" $(CUSTOMIZE_SOURCE_DIR)/*.js || true)
	if [ -n "$$file_hfroute_header" ]; then \
		hidden_hfroute_header='o().createElement("div",{className:c()("hfroute-header",{"has-free-banner":r,"has-custom-site-name":b}),style:{visibility:"hidden"}}'; \
		sed -i '' "s|$${hfroute_header}|$${hidden_hfroute_header}|g" "$$file_hfroute_header"; \
	fi

	# Hide loading screen icon
	echo ".LoadingAnimation__compass { display: none; }" >> $(CUSTOMIZE_SOURCE_DIR)/css/initial_loading_screen.css

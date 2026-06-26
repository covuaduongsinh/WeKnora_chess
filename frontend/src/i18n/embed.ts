import { createI18n } from 'vue-i18n'

const messages = {
  "vi-VN": {
    "embedPublish": {
      "title": "Nhúng web",
      "description": "Nhúng trợ lý AI vào trang web của bạn, khách truy cập có thể bắt đầu trò chuyện qua cửa sổ chat trong trang hoặc cửa sổ nổi ở góc dưới bên phải.",
      "create": "Tạo kênh nhúng",
      "empty": "Chưa có kênh nhúng",
      "unnamed": "Kênh chưa đặt tên",
      "agent": "Trợ lý AI",
      "rateLimit": "Giới hạn tần suất",
      "rateLimitUnit": "lần/phút",
      "allowedOrigins": "Danh sách trắng tên miền",
      "embedCode": "Mã nhúng",
      "widgetCode": "Script cửa sổ nổi",
      "copyCode": "Sao chép mã",
      "rotateToken": "Xoay Token",
      "delete": "Xóa",
      "edit": "Sửa",
      "createTitle": "Tạo kênh nhúng",
      "editTitle": "Sửa kênh nhúng",
      "name": "Tên",
      "namePlaceholder": "Ví dụ: CSKH website chính thức",
      "welcomeMessage": "Lời chào",
      "welcomePlaceholder": "Xin chào, tôi có thể giúp gì cho bạn?",
      "originsLabel": "Danh sách trắng tên miền (mỗi dòng một mục, để trống nghĩa là không giới hạn)",
      "originsPlaceholder": "https://shop.example.com",
      "rateLimitLabel": "Giới hạn yêu cầu mỗi phút",
      "debug": "Xem trước gỡ lỗi",
      "createdDebugHint": "Đã tạo kênh nhúng, có thể nhấn «Xem trước gỡ lỗi» để trải nghiệm ở tab mới",
      "primaryColor": "Màu chủ đạo",
      "pageTitle": "Tiêu đề trang",
      "pageTitlePlaceholder": "CSKH AI",
      "tokenHint": "Token chỉ hiển thị khi tạo hoặc xoay, vui lòng nhấn «Xoay Token» để lấy mã nhúng",
      "created": "Đã tạo kênh nhúng",
      "updated": "Đã cập nhật kênh nhúng",
      "deleted": "Đã xóa",
      "tokenRotated": "Đã xoay Token",
      "copied": "Đã sao chép mã nhúng",
      "loadError": "Tải thất bại",
      "missingChannel": "Thiếu kênh nhúng hoặc Token",
      "invalidChannel": "Kênh nhúng không hợp lệ",
      "sessionFailed": "Không thể tạo phiên trò chuyện, vui lòng thử lại sau",
      "channelDisabled": "Kênh nhúng đã bị tắt, vui lòng bật lại trong mục «Nhúng web» của trình chỉnh sửa Agent",
      "loading": "Đang tải...",
      "tabIframe": "iframe",
      "tabWidget": "Cửa sổ nổi",
      "widgetPosition": "Vị trí cửa sổ nổi",
      "widgetPreview": "Xem trước cửa sổ nổi",
      "positionBottomRight": "Góc dưới phải",
      "positionBottomLeft": "Góc dưới trái",
      "positionTopRight": "Góc trên phải",
      "positionTopLeft": "Góc trên trái",
      "publishToken": "Token đăng",
      "publishTokenHelp": "Token đăng (em_…) là khóa lâu dài của kênh nhúng, tương đương API Key. Chỉ hiển thị một lần khi tạo hoặc xoay, cần đưa vào script cửa sổ nổi của website bên thứ ba; giao diện danh sách không trả về để tránh lộ.",
      "sessionTokenHelp": "Sau khi khách mở chat, iframe sẽ dùng Token đăng để đổi lấy Token phiên ngắn hạn (ems_…, khoảng 30 phút), các yêu cầu sau dùng Token phiên, tránh để lộ Token đăng lâu dài trong URL.",
      "rotateTokenHelp": "Xoay sẽ hủy Token đăng cũ, mọi mã nhúng đã triển khai cần cập nhật đồng bộ, nếu không website bên thứ ba sẽ không truy cập được.",
      "revealToken": "Hiện",
      "hideToken": "Ẩn",
      "copyToken": "Sao chép Token",
      "tokenCopied": "Đã sao chép Token",
      "awaitingToken": "Đang chờ trang chủ cung cấp Token…",
      "preview": "Xem trước",
      "previewLoading": "Đang tải bản xem trước…",
      "previewIframeHint": "Mô phỏng hiệu ứng khi website bên thứ ba nhúng bằng iframe (giống mã sao chép).",
      "previewWidgetHint": "Mô phỏng hiệu ứng sau khi website bên thứ ba nạp script cửa sổ nổi; website thật sẽ do trang chủ truyền Token qua postMessage.",
      "previewMockPage": "Mô phỏng trang chủ",
      "defaultChatTitle": "Trợ lý AI",
      "newChat": "Cuộc trò chuyện mới",
      "rotateConfirmTitle": "Xác nhận xoay Token?",
      "rotateConfirmBody": "Sau khi xoay, Token cũ mất hiệu lực ngay lập tức, mọi mã nhúng đã đăng cần cập nhật toàn bộ.",
      "tokenRequiredForPreview": "Cần Token đăng mới xem trước được. Vui lòng tạo kênh trước, hoặc nhấn «Xoay Token» để lấy Token mới."
    },
    "chat": {
      "title": "Trò chuyện",
      "newChat": "Cuộc trò chuyện mới",
      "suggestedQuestions": "Bạn có thể hỏi tôi như vầy",
      "suggestedQuestionsLoading": "Đang tải câu hỏi gợi ý...",
      "inputPlaceholder": "Vui lòng nhập tin nhắn của bạn...",
      "send": "Gửi",
      "thinking": "Đang suy nghĩ...",
      "regenerate": "Tạo lại",
      "copy": "Sao chép",
      "delete": "Xóa",
      "reference": "Trích dẫn",
      "noMessages": "Chưa có tin nhắn",
      "waitingForAnswer": "Đang chờ câu trả lời...",
      "cannotAnswer": "Xin lỗi, tôi không thể trả lời câu hỏi này.",
      "summarizingAnswer": "Đang tóm tắt câu trả lời...",
      "loading": "Đang tải...",
      "enterDescription": "Nhập mô tả",
      "referencedContent": "Đã trích dẫn {count} tài liệu liên quan",
      "deepThinking": "Đã hoàn tất suy luận sâu",
      "knowledgeBaseQandA": "Hỏi đáp kho tri thức",
      "askKnowledgeBase": "Đặt câu hỏi cho kho tri thức",
      "sourcesCount": "{count} nguồn",
      "pleaseEnterContent": "Vui lòng nhập nội dung!",
      "pleaseUploadKnowledgeBase": "Vui lòng tải lên kho tri thức trước!",
      "replyingPleaseWait": "Đang trả lời, vui lòng thử lại sau!",
      "createSessionFailed": "Tạo phiên thất bại",
      "createSessionError": "Lỗi khi tạo phiên",
      "unableToGetKnowledgeBaseId": "Không thể lấy ID kho tri thức",
      "summaryInProgress": "Đang tóm tắt câu trả lời...",
      "thinkingAlt": "Đang suy nghĩ",
      "deepThoughtCompleted": "Đã suy luận sâu",
      "deepThoughtAlt": "Đã hoàn tất suy luận sâu",
      "referencesTitle": "Đã tham khảo {count} nội dung liên quan",
      "referencesDocCount": "Đã trích dẫn {count} tài liệu",
      "referencesDocAndWebCount": "Đã trích dẫn {docCount} tài liệu và {webCount} trang web",
      "referenceChunkCount": "{count} đoạn",
      "fallbackHint": "Không truy hồi được nội dung liên quan từ kho tri thức, trên đây là câu trả lời trực tiếp của mô hình",
      "requestInfoTitle": "Thông tin yêu cầu",
      "requestInfoRequestId": "Request ID",
      "requestInfoMessageId": "ID tin nhắn",
      "requestInfoSessionId": "ID phiên",
      "requestInfoUrl": "Yêu cầu",
      "requestInfoSentAt": "Thời gian gửi",
      "requestInfoEmpty": "Chưa có thông tin yêu cầu",
      "channelWeb": "Web",
      "channelApi": "API",
      "channelIm": "IM",
      "chunkLabel": "Đoạn {index}:",
      "navigateToDocument": "Xem chi tiết tài liệu",
      "referenceIconAlt": "Biểu tượng nội dung tham khảo",
      "chunkIdLabel": "ID đoạn:",
      "documentIdLabel": "ID tài liệu:",
      "faqIdLabel": "FAQ ID:",
      "faqContainerIdLabel": "ID tài liệu chứa:",
      "faqAnswersLabel": "Câu trả lời:",
      "chunkOrdinal": "Đoạn {index}",
      "previewContent": "Xem trước nội dung",
      "noPlanSteps": "Không cung cấp các bước cụ thể",
      "chunkIndexLabel": "Đoạn #{index}",
      "chunkPositionLabel": "(vị trí: {position})",
      "noRelatedChunks": "Không tìm thấy đoạn liên quan",
      "noSearchResults": "Không tìm thấy kết quả tìm kiếm",
      "relevanceHigh": "Liên quan cao",
      "relevanceMedium": "Liên quan vừa",
      "relevanceLow": "Liên quan thấp",
      "relevanceWeak": "Liên quan yếu",
      "webSearchNoResults": "Không tìm thấy kết quả tìm kiếm",
      "otherSource": "Nguồn khác",
      "webGroupIntro": "{count} nội dung dưới đây đến từ",
      "graphConfigTitle": "Cấu hình đồ thị",
      "entityTypesLabel": "Loại thực thể:",
      "relationTypesLabel": "Loại quan hệ:",
      "graphResultsHeader": "Tìm thấy {count} kết quả liên quan",
      "graphNoResults": "Không tìm thấy thông tin đồ thị liên quan",
      "unknownLink": "Liên kết không xác định",
      "contentLengthLabel": "Độ dài {value}",
      "notProvided": "Không cung cấp",
      "promptLabel": "Câu lệnh",
      "errorMessageLabel": "Thông báo lỗi",
      "summaryLabel": "Tóm tắt",
      "rawTextLabel": "Văn bản gốc",
      "collapseRaw": "Thu gọn bản gốc",
      "expandRaw": "Mở rộng bản gốc",
      "noWebContent": "Không lấy được nội dung trang web",
      "lengthChars": "{value} ký tự",
      "lengthThousands": "{value} nghìn ký tự",
      "lengthTenThousands": "{value} vạn ký tự",
      "sqlQueryExecuted": "Truy vấn SQL đã thực thi:",
      "sqlResultsLabel": "Kết quả trả về:",
      "rowsLabel": "hàng",
      "columnsLabel": "cột",
      "noDatabaseRecords": "Không tìm thấy bản ghi phù hợp",
      "nullValuePlaceholder": "<NULL>",
      "documentTitleLabel": "Tiêu đề tài liệu:",
      "chunkCountLabel": "Số đoạn:",
      "chunkCountValue": "{count} đoạn",
      "documentDescriptionLabel": "Mô tả tài liệu:",
      "documentStatusLabel": "Trạng thái xử lý:",
      "documentSourceLabel": "Nguồn:",
      "documentFileLabel": "Thông tin tệp:",
      "documentMetadataLabel": "Siêu dữ liệu",
      "documentInfoSummaryLabel": "Thông tin tài liệu",
      "documentInfoCount": "Thành công {count} / Yêu cầu {requested}",
      "documentInfoErrors": "Chi tiết lỗi",
      "documentInfoEmpty": "Chưa có thông tin tài liệu",
      "statusDescription": "Mô tả trạng thái",
      "statusIndexed": "Tài liệu đã được lập chỉ mục và có thể tìm kiếm",
      "statusSearchable": "Có thể dùng công cụ tìm kiếm để tra nội dung tài liệu",
      "statusChunkDetailAvailable": "Có thể dùng get_chunk_detail để xem chi tiết đoạn",
      "positionLabel": "Vị trí:",
      "chunkPositionValue": "Đoạn thứ {index}",
      "contentLengthLabelSimple": "Độ dài nội dung:",
      "fullContentLabel": "Nội dung đầy đủ",
      "copyContent": "Sao chép nội dung",
      "knowledgeBaseCount": "Tổng {count} kho tri thức",
      "noKnowledgeBases": "Không có kho tri thức khả dụng",
      "rawOutputLabel": "Kết quả gốc",
      "wikiWritePageTitle": "Ghi trang Wiki",
      "wikiReplaceTextTitle": "Thay văn bản Wiki",
      "wikiRenamePageTitle": "Đổi tên trang Wiki",
      "wikiDeletePageTitle": "Xóa trang Wiki",
      "wikiActionCreated": "Đã tạo",
      "wikiActionUpdated": "Đã cập nhật",
      "wikiActionRenamed": "Đã đổi tên",
      "wikiActionDeleted": "Đã xóa",
      "wikiFieldSlug": "Đường dẫn trang",
      "wikiFieldTitle": "Tiêu đề",
      "wikiFieldPageType": "Loại",
      "wikiFieldSummary": "Tóm tắt",
      "wikiFieldOldText": "Văn bản cũ",
      "wikiFieldNewText": "Văn bản mới",
      "wikiFieldOldSlug": "Đường dẫn cũ",
      "wikiFieldNewSlug": "Đường dẫn mới",
      "wikiFieldAffectedPages": "Trang bị ảnh hưởng",
      "wikiAffectedCount": "Liên kết của {count} trang đã được cập nhật",
      "selectKnowledgeBaseWarning": "Vui lòng chọn ít nhất một kho tri thức",
      "processError": "Xử lý gặp lỗi",
      "sessionExcerpt": "Trích đoạn phiên",
      "noAnswerContent": "(không có nội dung trả lời)",
      "noMatchFound": "Không tìm thấy nội dung phù hợp",
      "deleteSessionFailed": "Xóa thất bại, vui lòng thử lại sau!",
      "imageTooMany": "Tải lên tối đa 5 ảnh",
      "imageTypeSizeError": "Chỉ hỗ trợ định dạng JPG/PNG/GIF/WEBP, mỗi ảnh không quá 10MB",
      "imageReadFailed": "Đọc ảnh thất bại",
      "imageUploadTooltip": "Tải ảnh lên (hỗ trợ dán/kéo-thả)",
      "attachmentUploadTooltip": "Tải tệp đính kèm lên (tài liệu, âm thanh...)",
      "attachmentWithCount": "Đã tải lên {count} tệp đính kèm",
      "attachmentTooMany": "Tải lên tối đa {max} tệp đính kèm",
      "attachmentTooLarge": "Tệp {name} vượt giới hạn {max}MB",
      "attachmentTypeNotSupported": "Loại tệp không được hỗ trợ: {name}",
      "copySuccess": "Đã sao chép vào clipboard",
      "copyFailed": "Sao chép thất bại",
      "emptyContentWarning": "Câu trả lời hiện tại trống",
      "editorOpened": "Đã mở trình soạn thảo, vui lòng chọn kho tri thức rồi lưu"
    },
    "common": {
      "loading": "Đang tải...",
      "confirm": "Xác nhận",
      "cancel": "Hủy",
      "copy": "Sao chép",
      "copied": "Đã sao chép",
      "finish": "Hoàn tất"
    },
    "error": {
      "tokenNotFound": "Không tìm thấy token đăng nhập, vui lòng đăng nhập lại",
      "invalidImageLink": "Liên kết ảnh không hợp lệ",
      "streamFailed": "Kết nối luồng thất bại"
    },
    "agent": {
      "taskLabel": "Nhiệm vụ:",
      "think": "Suy nghĩ",
      "copy": "Sao chép",
      "addToKnowledgeBase": "Thêm vào kho tri thức",
      "updatePlan": "Cập nhật kế hoạch",
      "webSearchFound": "Tìm thấy <strong>{count}</strong> kết quả tìm kiếm trực tuyến",
      "argumentsLabel": "Tham số",
      "toolFallback": "Công cụ",
      "stepsCompleted": "Đã hoàn thành <strong>{steps}</strong> bước",
      "stepsCompletedWithDuration": "Đã hoàn thành <strong>{steps}</strong> bước, mất <strong>{duration}</strong>",
      "reasoningRounds": "Suy nghĩ <strong>{rounds}</strong> vòng",
      "toolCalls": "Gọi công cụ <strong>{tools}</strong> lần",
      "durationSuffix": "mất <strong>{duration}</strong>",
      "stepSummarySeparator": " · "
    },
    "agentStream": {
      "toolApproval": {
        "banner": "Công cụ MCP này đã được đánh dấu «cần duyệt thủ công», hãy xác nhận tham số rồi mới thực thi",
        "service": "Dịch vụ",
        "tool": "Công cụ",
        "argsLabel": "Tham số gọi",
        "argsModified": "Đã sửa",
        "countdown": "Còn khoảng {seconds} giây",
        "approve": "Duyệt và thực thi",
        "reject": "Từ chối",
        "approvedTag": "Đã duyệt",
        "rejectedTag": "Đã từ chối",
        "invalidJson": "Tham số không phải JSON hợp lệ",
        "submitted": "Đã gửi",
        "submitFailed": "Gửi thất bại",
        "userRejected": "Người dùng từ chối"
      },
      "tools": {
        "searchKnowledge": "Truy hồi kho tri thức",
        "grepChunks": "Tìm từ khóa",
        "webSearch": "Tìm kiếm trực tuyến",
        "webFetch": "Thu thập trang web",
        "getDocumentInfo": "Lấy thông tin tài liệu",
        "listKnowledgeChunks": "Xem các khối tri thức",
        "getRelatedDocuments": "Tìm tài liệu liên quan",
        "getDocumentContent": "Lấy nội dung tài liệu",
        "todoWrite": "Quản lý kế hoạch",
        "knowledgeGraphExtract": "Trích xuất đồ thị tri thức",
        "thinking": "Suy nghĩ",
        "imageAnalysis": "Xem nội dung ảnh",
        "queryUnderstand": "Hiểu câu hỏi",
        "queryKnowledgeGraph": "Truy vấn đồ thị tri thức",
        "readSkill": "Đọc kỹ năng",
        "executeSkillScript": "Chạy script kỹ năng",
        "dataAnalysis": "Phân tích dữ liệu",
        "dataSchema": "Cấu trúc dữ liệu",
        "databaseQuery": "Truy vấn cơ sở dữ liệu"
      },
      "summary": {
        "searchKb": "Truy hồi kho tri thức <strong>{count}</strong> lần",
        "thinking": "Suy nghĩ <strong>{count}</strong> lần",
        "callTool": "Gọi {name}",
        "callTools": "Gọi công cụ {names}",
        "intermediateSteps": "<strong>{count}</strong> bước trung gian",
        "separator": ", ",
        "comma": ", "
      },
      "citation": {
        "loading": "Đang tải...",
        "notFound": "Không tìm thấy nội dung",
        "loadFailed": "Tải thất bại",
        "chunkId": "ID đoạn",
        "noKbForWiki": "Không nhận diện được kho tri thức liên kết, không thể mở Wiki"
      },
      "toolSummary": {
        "getDocument": "Lấy tài liệu: {title}",
        "document": "Tài liệu",
        "listChunks": "Xem {title}",
        "listFaqEntry": "Xem FAQ: {question}",
        "deepThinking": "Suy luận sâu"
      },
      "plan": {
        "inProgress": "Đang thực hiện",
        "pending": "Chờ xử lý",
        "completed": "Hoàn tất"
      },
      "search": {
        "noResults": "Không tìm thấy nội dung phù hợp",
        "foundResultsFromFiles": "Tìm thấy {count} kết quả từ {files} tệp",
        "foundResults": "Tìm thấy {count} kết quả",
        "webResults": "Tìm thấy {count} kết quả tìm kiếm trực tuyến",
        "grepSummary": "Tìm thấy {chunks} đoạn khớp từ {docs} tài liệu"
      },
      "grepResults": {
        "chunkHits": "{count} đoạn",
        "keywordHits": "{count} lần",
        "titleMatch": "Khớp tiêu đề",
        "faqEntry": "Mục FAQ"
      },
      "knowledgeChunksList": {
        "chunkRange": "Đã tải {fetched} / {total} khối",
        "page": "Trang {page}, mỗi trang {pageSize}"
      },
      "ragPipeline": {
        "understanding": "Đang hiểu câu hỏi...",
        "understandDone": "Đã hiểu xong câu hỏi",
        "searching": "Đang truy hồi kho tri thức...",
        "searchingWithQuery": "Đang truy hồi kho tri thức: «{query}»",
        "searchDone": "Truy hồi hoàn tất",
        "searchDoneWithQuery": "Truy hồi kho tri thức: «{query}»",
        "referencedDocs": "Trích dẫn <strong>{count}</strong> tài liệu",
        "referencedWebs": "Trích dẫn <strong>{count}</strong> trang web",
        "referencedDocAndWeb": "Trích dẫn <strong>{docCount}</strong> tài liệu và <strong>{webCount}</strong> trang web"
      },
      "toolStatus": {
        "calling": "Đang gọi {name}...",
        "searchKb": "Truy hồi kho tri thức",
        "searchKbFailed": "Truy hồi kho tri thức thất bại",
        "webSearch": "Tìm kiếm trực tuyến",
        "webSearchFailed": "Tìm kiếm trực tuyến thất bại",
        "grepSearch": "Tìm từ khóa",
        "grepSearchFailed": "Tìm từ khóa thất bại",
        "getDocInfo": "Lấy thông tin tài liệu",
        "getDocInfoFailed": "Lấy thông tin tài liệu thất bại",
        "viewDocument": "Xem tài liệu",
        "thinkingDone": "Suy nghĩ xong",
        "thinkingFailed": "Suy nghĩ thất bại",
        "updateTodos": "Cập nhật danh sách nhiệm vụ",
        "updateTodosFailed": "Cập nhật danh sách nhiệm vụ thất bại",
        "imageAnalyzing": "Đang xem nội dung ảnh...",
        "imageAnalysisDone": "Đã xem nội dung ảnh",
        "imageAnalysisFailed": "Xem nội dung ảnh thất bại",
        "queryUnderstanding": "Đang hiểu câu hỏi...",
        "queryUnderstandDone": "Đã hiểu xong câu hỏi",
        "called": "Gọi {name}",
        "calledFailed": "Gọi {name} thất bại"
      },
      "copy": {
        "emptyContent": "Câu trả lời hiện tại trống, không thể sao chép",
        "success": "Đã sao chép vào clipboard",
        "failed": "Sao chép thất bại, vui lòng sao chép thủ công"
      },
      "saveToKb": {
        "emptyContent": "Câu trả lời hiện tại trống, không thể lưu vào kho tri thức",
        "editorOpened": "Đã mở trình soạn thảo, vui lòng chọn kho tri thức rồi lưu"
      }
    },
    "input": {
      "placeholder": "Hỏi trực tiếp mô hình",
      "stopGeneration": "Dừng sinh",
      "send": "Gửi",
      "webSearch": {
        "label": "Tìm kiếm trực tuyến",
        "toggleOn": "Bật tìm kiếm trực tuyến",
        "toggleOff": "Tắt tìm kiếm trực tuyến",
        "agentDisabled": "Trợ lý AI hiện tại chưa bật tìm kiếm trực tuyến"
      },
      "imageUpload": {
        "label": "Tải ảnh lên",
        "tooltip": "Tải ảnh lên",
        "agentDisabled": "Trợ lý AI hiện tại chưa bật tải ảnh lên"
      },
      "fileUpload": {
        "label": "Tải tệp đính kèm lên",
        "tooltip": "Tải tệp đính kèm như tài liệu",
        "tooMany": "Tải lên tối đa 5 tệp đính kèm",
        "tooLarge": "Tệp đính kèm vượt giới hạn 20MB"
      },
      "messages": {
        "enterContent": "Vui lòng nhập nội dung trước!",
        "selectKnowledge": "Vui lòng chọn kho tri thức trước!",
        "replying": "Đang trả lời, vui lòng thử lại sau!",
        "agentSwitchedOn": "Đã chuyển sang suy luận thông minh",
        "agentSwitchedOff": "Đã chuyển sang hỏi đáp nhanh",
        "agentSelected": "Đã chọn trợ lý AI «{name}»",
        "agentEnabled": "Đã bật chế độ Agent",
        "agentDisabled": "Đã tắt chế độ Agent",
        "agentNotReadyDetail": "Agent chưa sẵn sàng, cần cấu hình các mục sau: {reasons}",
        "webSearchNotConfigured": "Chưa cấu hình công cụ tìm kiếm trực tuyến, vui lòng hoàn tất chọn công cụ và cấu hình giao diện trong cài đặt trước.",
        "webSearchEnabled": "Đã bật tìm kiếm trực tuyến",
        "webSearchDisabled": "Đã tắt tìm kiếm trực tuyến",
        "sessionMissing": "ID phiên không tồn tại",
        "messageMissing": "Không lấy được ID tin nhắn, vui lòng làm mới trang rồi thử lại",
        "stopSuccess": "Đã dừng sinh",
        "stopFailed": "Dừng thất bại, vui lòng thử lại"
      }
    },
    "knowledgeEditor": {
      "wikiBrowser": {
        "viewInGraph": "Xem trong đồ thị",
        "version": "v{ver}",
        "filterSummary": "Tóm tắt",
        "filterEntity": "Thực thể",
        "filterConcept": "Khái niệm",
        "filterSynthesis": "Tổng hợp",
        "filterComparison": "So sánh"
      }
    }
  },
  "en-US": {
    "embedPublish": {
      "title": "Web Page Embed",
      "description": "Embed this agent on your website so visitors can chat via an in-page window or a floating launcher.",
      "create": "New embed channel",
      "empty": "No embed channels yet",
      "unnamed": "Unnamed channel",
      "agent": "Agent",
      "rateLimit": "Rate limit",
      "rateLimitUnit": "/min",
      "allowedOrigins": "Allowed origins",
      "embedCode": "Embed code",
      "widgetCode": "Widget script",
      "copyCode": "Copy code",
      "rotateToken": "Rotate token",
      "delete": "Delete",
      "edit": "Edit",
      "createTitle": "New embed channel",
      "editTitle": "Edit embed channel",
      "name": "Name",
      "namePlaceholder": "e.g. Website support",
      "welcomeMessage": "Welcome message",
      "welcomePlaceholder": "Hi! How can I help you?",
      "originsLabel": "Allowed origins (one per line, empty = allow all)",
      "originsPlaceholder": "https://shop.example.com",
      "rateLimitLabel": "Requests per minute",
      "debug": "Debug preview",
      "createdDebugHint": "Embed channel created — use Debug preview to open it in a new tab",
      "primaryColor": "Primary color",
      "pageTitle": "Page title",
      "pageTitlePlaceholder": "AI Assistant",
      "tokenHint": "Token is shown only on create or rotate — click Rotate token to get embed code",
      "created": "Embed channel created",
      "updated": "Embed channel updated",
      "deleted": "Deleted",
      "tokenRotated": "Token rotated",
      "copied": "Embed code copied",
      "loadError": "Failed to load",
      "missingChannel": "Missing embed channel or token",
      "invalidChannel": "Invalid embed channel",
      "sessionFailed": "Failed to create chat session, please try again",
      "channelDisabled": "This embed channel is disabled. Re-enable it under Agent editor → Web Page Embed",
      "loading": "Loading...",
      "tabIframe": "iframe",
      "tabWidget": "Widget",
      "widgetPosition": "Widget position",
      "widgetPreview": "Widget preview",
      "positionBottomRight": "Bottom right",
      "positionBottomLeft": "Bottom left",
      "positionTopRight": "Top right",
      "positionTopLeft": "Top left",
      "publishToken": "Publish token",
      "publishTokenHelp": "The publish token (em_…) is a long-lived secret for this embed channel—like an API key. It is shown only when created or rotated and must be placed in the host site widget script; list APIs never return it.",
      "sessionTokenHelp": "After a visitor opens chat, the iframe exchanges the publish token for a short-lived session token (ems_…, ~30 min). Later API calls use the session token so the publish token is not kept in the URL.",
      "rotateTokenHelp": "Rotating invalidates the previous publish token. Every deployed embed snippet must be updated or third-party sites will lose access.",
      "revealToken": "Reveal",
      "hideToken": "Hide",
      "copyToken": "Copy token",
      "tokenCopied": "Token copied",
      "awaitingToken": "Waiting for host page to provide token…",
      "preview": "Preview",
      "previewLoading": "Loading preview…",
      "previewIframeHint": "Shows how the iframe embed looks on a third-party page (same as the copied snippet).",
      "previewWidgetHint": "Shows the floating widget on a mock host page. On a real site the host passes the token via postMessage.",
      "previewMockPage": "Mock host page",
      "defaultChatTitle": "AI Assistant",
      "newChat": "New chat",
      "rotateConfirmTitle": "Rotate publish token?",
      "rotateConfirmBody": "The old token stops working immediately. Update every deployed embed snippet.",
      "tokenRequiredForPreview": "A publish token is required to preview. Create a channel or rotate the token first."
    },
    "chat": {
      "title": "Chat",
      "newChat": "New Chat",
      "suggestedQuestions": "You can ask me",
      "suggestedQuestionsLoading": "Loading suggestions...",
      "inputPlaceholder": "Enter your message...",
      "send": "Send",
      "thinking": "Thinking...",
      "regenerate": "Regenerate",
      "copy": "Copy",
      "delete": "Delete",
      "reference": "Reference",
      "noMessages": "No messages",
      "waitingForAnswer": "Waiting for answer...",
      "cannotAnswer": "Sorry, I cannot answer this question.",
      "summarizingAnswer": "Summarizing answer...",
      "loading": "Loading...",
      "referencedContent": "{count} related materials used",
      "deepThinking": "Deep thinking completed",
      "knowledgeBaseQandA": "Knowledge Base Q&A",
      "askKnowledgeBase": "Ask the knowledge base",
      "sourcesCount": "{count} sources",
      "pleaseEnterContent": "Please enter content!",
      "pleaseUploadKnowledgeBase": "Please upload knowledge base first!",
      "replyingPleaseWait": "Replying, please try again later!",
      "createSessionFailed": "Failed to create session",
      "createSessionError": "Session creation error",
      "unableToGetKnowledgeBaseId": "Unable to get knowledge base ID",
      "summaryInProgress": "Summarizing answer…",
      "thinkingAlt": "Thinking in progress",
      "deepThoughtCompleted": "Deep thinking completed",
      "deepThoughtAlt": "Deep thinking finished",
      "referencesTitle": "Referenced {count} related item(s)",
      "referencesDocCount": "Referenced {count} document(s)",
      "referencesDocAndWebCount": "Referenced {docCount} document(s) and {webCount} web page(s)",
      "referenceChunkCount": "{count} chunk(s)",
      "fallbackHint": "No relevant content found in knowledge base. Above is a direct response from the model.",
      "requestInfoTitle": "Request info",
      "requestInfoRequestId": "Request ID",
      "requestInfoMessageId": "Message ID",
      "requestInfoSessionId": "Session ID",
      "requestInfoUrl": "Request",
      "requestInfoSentAt": "Sent at",
      "requestInfoEmpty": "No request info available",
      "channelWeb": "Web",
      "channelApi": "API",
      "channelIm": "IM",
      "chunkLabel": "Chunk {index}:",
      "navigateToDocument": "View document details",
      "referenceIconAlt": "Reference materials icon",
      "chunkIdLabel": "Chunk ID:",
      "documentIdLabel": "Document ID:",
      "faqIdLabel": "FAQ ID:",
      "faqContainerIdLabel": "Container ID:",
      "faqAnswersLabel": "Answers:",
      "chunkOrdinal": "Chunk {index}",
      "previewContent": "Preview content",
      "noPlanSteps": "No detailed steps provided",
      "chunkIndexLabel": "Chunk #{index}",
      "chunkPositionLabel": "(Position: {position})",
      "noRelatedChunks": "No related chunks found",
      "noSearchResults": "No search results found",
      "relevanceHigh": "High relevance",
      "relevanceMedium": "Medium relevance",
      "relevanceLow": "Low relevance",
      "relevanceWeak": "Weak relevance",
      "webSearchNoResults": "No web search results found",
      "otherSource": "Other sources",
      "webGroupIntro": "The following {count} items are from",
      "graphConfigTitle": "Graph Configuration",
      "entityTypesLabel": "Entity types:",
      "relationTypesLabel": "Relation types:",
      "graphResultsHeader": "{count} related results found",
      "graphNoResults": "No related graph information found",
      "unknownLink": "Unknown link",
      "contentLengthLabel": "Length {value}",
      "notProvided": "Not provided",
      "promptLabel": "Prompt",
      "errorMessageLabel": "Error message",
      "summaryLabel": "Summary",
      "rawTextLabel": "Raw text",
      "collapseRaw": "Collapse original",
      "expandRaw": "Expand original",
      "noWebContent": "No web content fetched",
      "lengthChars": "{value} characters",
      "lengthThousands": "{value}k characters",
      "lengthTenThousands": "{value} ten-thousand characters",
      "sqlQueryExecuted": "Executed SQL query:",
      "sqlResultsLabel": "Results:",
      "rowsLabel": "rows",
      "columnsLabel": "columns",
      "noDatabaseRecords": "No matching records found",
      "nullValuePlaceholder": "<NULL>",
      "documentTitleLabel": "Document title:",
      "chunkCountLabel": "Chunk count:",
      "chunkCountValue": "{count} chunks",
      "documentDescriptionLabel": "Description:",
      "documentStatusLabel": "Status:",
      "documentSourceLabel": "Source:",
      "documentFileLabel": "File:",
      "documentMetadataLabel": "Metadata",
      "documentInfoSummaryLabel": "Document info",
      "documentInfoCount": "{count} of {requested} documents retrieved",
      "documentInfoErrors": "Errors",
      "documentInfoEmpty": "No document information available",
      "statusDescription": "Status notes",
      "statusIndexed": "Document is indexed and searchable",
      "statusSearchable": "Search tools can locate document content",
      "statusChunkDetailAvailable": "Use get_chunk_detail to view chunk details",
      "positionLabel": "Position:",
      "chunkPositionValue": "Chunk #{index}",
      "contentLengthLabelSimple": "Content length:",
      "fullContentLabel": "Full content",
      "copyContent": "Copy content",
      "knowledgeBaseCount": "{count} knowledge bases",
      "noKnowledgeBases": "No knowledge bases available",
      "enterDescription": "Enter description",
      "rawOutputLabel": "Raw output",
      "wikiWritePageTitle": "Wiki Page Write",
      "wikiReplaceTextTitle": "Wiki Text Replace",
      "wikiRenamePageTitle": "Wiki Page Rename",
      "wikiDeletePageTitle": "Wiki Page Delete",
      "wikiActionCreated": "Created",
      "wikiActionUpdated": "Updated",
      "wikiActionRenamed": "Renamed",
      "wikiActionDeleted": "Deleted",
      "wikiFieldSlug": "Slug",
      "wikiFieldTitle": "Title",
      "wikiFieldPageType": "Type",
      "wikiFieldSummary": "Summary",
      "wikiFieldOldText": "Old text",
      "wikiFieldNewText": "New text",
      "wikiFieldOldSlug": "Old slug",
      "wikiFieldNewSlug": "New slug",
      "wikiFieldAffectedPages": "Affected pages",
      "wikiAffectedCount": "{count} page link(s) updated",
      "selectKnowledgeBaseWarning": "Please select at least one knowledge base",
      "processError": "Processing error",
      "sessionExcerpt": "Session Excerpt",
      "noAnswerContent": "(No answer content)",
      "noMatchFound": "No matching content found",
      "deleteSessionFailed": "Delete failed, please try again later!",
      "imageTooMany": "Maximum 5 images allowed",
      "imageTypeSizeError": "Only JPG/PNG/GIF/WEBP under 10MB supported",
      "imageReadFailed": "Failed to read image",
      "imageUploadTooltip": "Upload image (paste/drop supported)",
      "attachmentUploadTooltip": "Upload attachment (documents, audio, etc.)",
      "attachmentWithCount": "{count} attachment(s) uploaded",
      "attachmentTooMany": "Maximum {max} attachments allowed",
      "attachmentTooLarge": "File {name} exceeds {max}MB limit",
      "attachmentTypeNotSupported": "Unsupported file type: {name}",
      "copySuccess": "Copied to clipboard",
      "copyFailed": "Copy failed",
      "emptyContentWarning": "Content is empty",
      "editorOpened": "Editor opened, please select a knowledge base and save"
    },
    "common": {
      "loading": "Loading...",
      "confirm": "Confirm",
      "cancel": "Cancel",
      "copy": "Copy",
      "copied": "Copied",
      "finish": "Finish"
    },
    "error": {
      "tokenNotFound": "Login token not found, please log in again",
      "invalidImageLink": "Invalid image link",
      "streamFailed": "Stream connection failed"
    },
    "agent": {
      "taskLabel": "Task:",
      "think": "Thinking",
      "copy": "Copy",
      "addToKnowledgeBase": "Add to Knowledge Base",
      "updatePlan": "Update Plan",
      "webSearchFound": "Found <strong>{count}</strong> web search result(s)",
      "argumentsLabel": "Arguments",
      "toolFallback": "Tool",
      "stepsCompleted": "Completed <strong>{steps}</strong> step(s)",
      "stepsCompletedWithDuration": "Completed <strong>{steps}</strong> step(s) in <strong>{duration}</strong>",
      "reasoningRounds": "<strong>{rounds}</strong> reasoning round(s)",
      "toolCalls": "<strong>{tools}</strong> tool call(s)",
      "durationSuffix": "<strong>{duration}</strong>",
      "stepSummarySeparator": " · "
    },
    "agentStream": {
      "toolApproval": {
        "banner": "This MCP tool requires human approval. Review parameters before execution.",
        "service": "Service",
        "tool": "Tool",
        "argsLabel": "Arguments",
        "argsModified": "Modified",
        "countdown": "About {seconds}s remaining",
        "approve": "Approve & run",
        "reject": "Reject",
        "approvedTag": "Approved",
        "rejectedTag": "Rejected",
        "invalidJson": "Arguments must be valid JSON",
        "submitted": "Submitted",
        "submitFailed": "Submit failed",
        "userRejected": "User rejected"
      },
      "tools": {
        "searchKnowledge": "Knowledge Search",
        "grepChunks": "Text Pattern Search",
        "webSearch": "Web Search",
        "webFetch": "Web Fetch",
        "getDocumentInfo": "Get Document Info",
        "listKnowledgeChunks": "List Knowledge Chunks",
        "getRelatedDocuments": "Find Related Documents",
        "getDocumentContent": "Get Document Content",
        "todoWrite": "Plan Management",
        "knowledgeGraphExtract": "Knowledge Graph Extraction",
        "thinking": "Thinking",
        "imageAnalysis": "Image Analysis",
        "queryUnderstand": "Understand Query",
        "queryKnowledgeGraph": "Knowledge Graph Query",
        "readSkill": "Read Skill",
        "executeSkillScript": "Execute Skill Script",
        "dataAnalysis": "Data Analysis",
        "dataSchema": "Data Schema",
        "databaseQuery": "Database Query"
      },
      "summary": {
        "searchKb": "Searched knowledge base <strong>{count}</strong> time(s)",
        "thinking": "Thought <strong>{count}</strong> time(s)",
        "callTool": "Called {name}",
        "callTools": "Called tools {names}",
        "intermediateSteps": "<strong>{count}</strong> intermediate step(s)",
        "separator": ", ",
        "comma": ", "
      },
      "citation": {
        "loading": "Loading...",
        "notFound": "Content not found",
        "loadFailed": "Failed to load",
        "chunkId": "Chunk ID",
        "noKbForWiki": "Unable to identify associated knowledge base. Cannot open Wiki."
      },
      "toolSummary": {
        "getDocument": "Get document: {title}",
        "document": "Document",
        "listChunks": "View {title}",
        "listFaqEntry": "View FAQ: {question}",
        "deepThinking": "Deep Thinking"
      },
      "plan": {
        "inProgress": "In Progress",
        "pending": "Pending",
        "completed": "Completed"
      },
      "search": {
        "noResults": "No matching content found",
        "foundResultsFromFiles": "Found {count} result(s) from {files} file(s)",
        "foundResults": "Found {count} result(s)",
        "webResults": "Found {count} web search result(s)",
        "grepSummary": "Found {chunks} matching chunk(s) across {docs} document(s)"
      },
      "grepResults": {
        "chunkHits": "{count} chunks",
        "keywordHits": "{count} hits",
        "titleMatch": "title",
        "faqEntry": "FAQ entry"
      },
      "knowledgeChunksList": {
        "chunkRange": "Loaded {fetched} / {total} chunks",
        "page": "Page {page}, {pageSize} per page"
      },
      "ragPipeline": {
        "understanding": "Understanding query...",
        "understandDone": "Query understood",
        "searching": "Searching knowledge base...",
        "searchingWithQuery": "Searching knowledge base: \"{query}\"",
        "searchDone": "Search complete",
        "searchDoneWithQuery": "Searched knowledge base: \"{query}\"",
        "referencedDocs": "Cited <strong>{count}</strong> documents",
        "referencedWebs": "Cited <strong>{count}</strong> web results",
        "referencedDocAndWeb": "Cited <strong>{docCount}</strong> documents and <strong>{webCount}</strong> web results"
      },
      "toolStatus": {
        "calling": "Calling {name}...",
        "searchKb": "Searching knowledge base",
        "searchKbFailed": "Knowledge base search failed",
        "webSearch": "Web search",
        "webSearchFailed": "Web search failed",
        "grepSearch": "Keyword search",
        "grepSearchFailed": "Keyword search failed",
        "getDocInfo": "Getting document info",
        "getDocInfoFailed": "Failed to get document info",
        "viewDocument": "View document",
        "thinkingDone": "Thinking complete",
        "thinkingFailed": "Thinking failed",
        "updateTodos": "Updating task list",
        "updateTodosFailed": "Failed to update task list",
        "imageAnalyzing": "Viewing image content...",
        "imageAnalysisDone": "Image content viewed",
        "imageAnalysisFailed": "Image viewing failed",
        "queryUnderstanding": "Understanding query...",
        "queryUnderstandDone": "Query understood",
        "called": "Called {name}",
        "calledFailed": "Failed to call {name}"
      },
      "copy": {
        "emptyContent": "Current response is empty, cannot copy",
        "success": "Copied to clipboard",
        "failed": "Copy failed, please copy manually"
      },
      "saveToKb": {
        "emptyContent": "Current response is empty, cannot save to knowledge base",
        "editorOpened": "Editor opened, please select a knowledge base and save"
      }
    },
    "input": {
      "placeholder": "Ask questions directly to the model",
      "stopGeneration": "Stop Generation",
      "send": "Send",
      "webSearch": {
        "label": "Web search",
        "toggleOn": "Enable web search",
        "toggleOff": "Disable web search",
        "agentDisabled": "Web search is not enabled for this agent"
      },
      "imageUpload": {
        "label": "Upload image",
        "tooltip": "Upload image",
        "agentDisabled": "Image upload is not enabled for this agent"
      },
      "fileUpload": {
        "label": "Upload file",
        "tooltip": "Upload document attachments",
        "tooMany": "Maximum 5 attachments",
        "tooLarge": "Attachment exceeds 20MB limit"
      },
      "messages": {
        "enterContent": "Please enter content first!",
        "selectKnowledge": "Please select a knowledge base first!",
        "replying": "Currently replying, please try again later!",
        "agentSwitchedOn": "Switched to Intelligent Reasoning",
        "agentSwitchedOff": "Switched to Quick Q&A",
        "agentSelected": "Selected agent \"{name}\"",
        "agentEnabled": "Agent Mode enabled",
        "agentDisabled": "Agent Mode disabled",
        "agentNotReadyDetail": "Agent is not ready. Please configure the following: {reasons}",
        "webSearchNotConfigured": "Web search engine is not configured. Please configure a provider and credentials in settings.",
        "webSearchEnabled": "Web search enabled",
        "webSearchDisabled": "Web search disabled",
        "sessionMissing": "Session ID does not exist",
        "messageMissing": "Unable to get message ID. Please refresh the page and try again.",
        "stopSuccess": "Generation stopped",
        "stopFailed": "Failed to stop. Please try again."
      }
    },
    "knowledgeEditor": {
      "wikiBrowser": {
        "viewInGraph": "View in Graph",
        "version": "v{ver}",
        "filterSummary": "Summaries",
        "filterEntity": "Entities",
        "filterConcept": "Concepts",
        "filterSynthesis": "Synthesis",
        "filterComparison": "Comparisons"
      }
    }
  }
} as const

type MessageTree = Record<string, unknown>

function deepMerge<T extends MessageTree>(base: T, patch: MessageTree): T {
  const out: MessageTree = { ...base }
  for (const key of Object.keys(patch)) {
    const patchVal = patch[key]
    const baseVal = base[key]
    if (
      patchVal &&
      typeof patchVal === 'object' &&
      !Array.isArray(patchVal) &&
      baseVal &&
      typeof baseVal === 'object' &&
      !Array.isArray(baseVal)
    ) {
      out[key] = deepMerge(baseVal as MessageTree, patchVal as MessageTree)
    } else {
      out[key] = patchVal
    }
  }
  return out as T
}

const koEmbedPublish = {
  embedPublish: {
    title: '웹 페이지 임베드',
    description: '에이전트를 웹 페이지에 임베드하여 방문자가 페이지 내 채팅창 또는 플로팅 버튼으로 대화할 수 있게 합니다.',
    create: '새 임베드 채널',
    empty: '임베드 채널 없음',
    unnamed: '이름 없는 채널',
    loading: '로딩 중...',
    awaitingToken: '호스트 페이지에서 토큰 제공 대기 중…',
    defaultChatTitle: 'AI 어시스턴트',
    newChat: '새 대화',
    preview: '미리보기',
    previewIframeHint: 'iframe 임베드가 외부 페이지에서 어떻게 보이는지 시뮬레이션합니다.',
    previewWidgetHint: '모의 호스트 페이지에서 플로팅 위젯을 표시합니다.',
    previewMockPage: '모의 호스트 페이지',
    previewLoading: '미리보기 로딩 중…',
    channelDisabled: '임베드 채널이 비활성화되었습니다. 에이전트 편집기 → 웹 페이지 임베드에서 다시 활성화하세요',
    invalidChannel: '잘못된 임베드 채널',
    sessionFailed: '대화 세션을 생성할 수 없습니다. 나중에 다시 시도하세요',
    missingChannel: '임베드 채널 또는 토큰 없음',
    loadError: '로드 실패',
  },
  common: {
    loading: '로딩 중...',
    confirm: '확인',
    cancel: '취소',
    copy: '복사',
    copied: '복사됨',
  },
  error: {
    tokenNotFound: '로그인 토큰을 찾을 수 없습니다. 다시 로그인하세요',
    invalidImageLink: '잘못된 이미지 링크',
    streamFailed: '스트림 연결 실패',
  },
  chat: {
    suggestedQuestions: '이렇게 물어보세요',
    imageTooMany: '이미지는 최대 5장까지 업로드할 수 있습니다',
    imageTypeSizeError: 'JPG/PNG/GIF/WEBP만 지원하며, 각 파일은 10MB 이하여야 합니다',
    imageReadFailed: '이미지를 읽지 못했습니다',
  },
  input: {
    placeholder: '모델에 직접 질문하세요',
    stopGeneration: '생성 중지',
    send: '보내기',
    webSearch: {
      label: '웹 검색',
      toggleOn: '웹 검색 켜기',
      toggleOff: '웹 검색 끄기',
      agentDisabled: '이 에이전트에서는 웹 검색이 활성화되지 않았습니다',
    },
    imageUpload: {
      label: '이미지 업로드',
      tooltip: '이미지 업로드',
      agentDisabled: '이 에이전트에서는 이미지 업로드가 활성화되지 않았습니다',
    },
    messages: {
      webSearchEnabled: '웹 검색이 켜졌습니다',
      webSearchDisabled: '웹 검색이 꺼졌습니다',
      stopSuccess: '생성이 중지되었습니다',
      stopFailed: '중지에 실패했습니다. 다시 시도하세요',
    },
  },
} as const

const ruEmbedPublish = {
  embedPublish: {
    title: 'Встраивание на веб-страницу',
    description: 'Встройте агента на свою веб-страницу: посетители смогут общаться через встроенное окно чата или плавающую кнопку.',
    create: 'Новый канал встраивания',
    empty: 'Каналов встраивания пока нет',
    unnamed: 'Без названия',
    loading: 'Загрузка...',
    awaitingToken: 'Ожидание токена от страницы-хоста…',
    defaultChatTitle: 'AI-ассистент',
    newChat: 'Новый чат',
    preview: 'Предпросмотр',
    previewIframeHint: 'Как iframe выглядит на сторонней странице.',
    previewWidgetHint: 'Плавающий виджет на mock-странице.',
    previewMockPage: 'Mock-страница хоста',
    previewLoading: 'Загрузка предпросмотра…',
    channelDisabled: 'Канал встраивания отключён. Включите в редакторе агента → Встраивание на веб-страницу',
    invalidChannel: 'Недействительный канал встраивания',
    sessionFailed: 'Не удалось создать сессию чата, попробуйте позже',
    missingChannel: 'Отсутствует канал встраивания или токен',
    loadError: 'Не удалось загрузить',
  },
  common: {
    loading: 'Загрузка...',
    confirm: 'Подтвердить',
    cancel: 'Отмена',
    copy: 'Копировать',
    copied: 'Скопировано',
  },
  error: {
    tokenNotFound: 'Токен входа не найден, войдите снова',
    invalidImageLink: 'Недействительная ссылка на изображение',
    streamFailed: 'Ошибка потокового соединения',
  },
  chat: {
    suggestedQuestions: 'Вы можете спросить так',
    imageTooMany: 'Можно загрузить не более 5 изображений',
    imageTypeSizeError: 'Поддерживаются только JPG/PNG/GIF/WEBP, каждый файл до 10 МБ',
    imageReadFailed: 'Не удалось прочитать изображение',
  },
  input: {
    placeholder: 'Задайте вопрос модели',
    stopGeneration: 'Остановить генерацию',
    send: 'Отправить',
    webSearch: {
      label: 'Веб-поиск',
      toggleOn: 'Включить веб-поиск',
      toggleOff: 'Выключить веб-поиск',
      agentDisabled: 'Веб-поиск не включён для этого агента',
    },
    imageUpload: {
      label: 'Загрузить изображение',
      tooltip: 'Загрузить изображение',
      agentDisabled: 'Загрузка изображений не включена для этого агента',
    },
    messages: {
      webSearchEnabled: 'Веб-поиск включён',
      webSearchDisabled: 'Веб-поиск выключен',
      stopSuccess: 'Генерация остановлена',
      stopFailed: 'Не удалось остановить. Попробуйте снова.',
    },
  },
} as const

const SUPPORTED_LOCALES = ['vi-VN', 'en-US', 'ko-KR', 'ru-RU'] as const
export type EmbedLocale = (typeof SUPPORTED_LOCALES)[number]

/** Isolated from the main app `locale` key so embed preview never hijacks admin UI language. */
export const EMBED_LOCALE_STORAGE_KEY = 'weknora-embed-locale'

/** Map host-provided locale strings to a supported embed locale tag. */
export function normalizeEmbedLocale(raw: string): EmbedLocale {
  const s = raw.trim().toLowerCase()
  if (s.startsWith('en')) return 'en-US'
  if (s.startsWith('ko')) return 'ko-KR'
  if (s.startsWith('ru')) return 'ru-RU'
  if (s.startsWith('vi')) return 'vi-VN'
  const exact = SUPPORTED_LOCALES.find((l) => l.toLowerCase() === s)
  return exact || 'vi-VN'
}

export function readEmbedLocaleFromUrl(): string {
  if (typeof window === 'undefined') return ''
  return new URLSearchParams(window.location.search).get('locale')?.trim() || ''
}

function resolveBrowserEmbedLocale(): EmbedLocale {
  const nav = typeof navigator !== 'undefined' ? navigator.language : ''
  return nav ? normalizeEmbedLocale(nav) : 'vi-VN'
}

function resolveInitialEmbedLocale(): EmbedLocale {
  const fromUrl = readEmbedLocaleFromUrl()
  if (fromUrl) return normalizeEmbedLocale(fromUrl)

  try {
    const saved = typeof localStorage !== 'undefined'
      ? localStorage.getItem(EMBED_LOCALE_STORAGE_KEY)
      : null
    if (saved) return normalizeEmbedLocale(saved)
  } catch {
    // localStorage may be unavailable in private mode.
  }

  return resolveBrowserEmbedLocale()
}

const locale = resolveInitialEmbedLocale()

const i18n = createI18n({
  legacy: false,
  locale,
  fallbackLocale: 'en-US',
  globalInjection: true,
  warnHtmlMessage: false,
  messages: {
    'vi-VN': messages['vi-VN'],
    'en-US': messages['en-US'],
    'ko-KR': deepMerge(messages['en-US'], koEmbedPublish),
    'ru-RU': deepMerge(messages['en-US'], ruEmbedPublish),
  },
})

type LocaleRef = { value: string }

/** Apply locale for the embed surface (isolated storage + optional active vue-i18n ref). */
export function applyEmbedLocale(raw: string, localeRef?: LocaleRef) {
  const next = normalizeEmbedLocale(raw)
  try {
    localStorage.setItem(EMBED_LOCALE_STORAGE_KEY, next)
  } catch {
    // localStorage may be unavailable in private mode.
  }
  if (localeRef) {
    localeRef.value = next
  } else {
    i18n.global.locale.value = next
  }
}

/** Honor `?locale=` on the embed URL for the currently mounted vue-i18n instance. */
export function syncEmbedLocaleFromUrl(localeRef: LocaleRef): boolean {
  const fromUrl = readEmbedLocaleFromUrl()
  if (!fromUrl) return false
  applyEmbedLocale(fromUrl, localeRef)
  return true
}

export default i18n
